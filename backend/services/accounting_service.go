package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"contaduria/mvp/shared"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type PocketBaseApp interface {
	FindCollectionByNameOrId(nameOrID string) (*core.Collection, error)
	FindAllRecords(collection any, exprs ...dbx.Expression) ([]*core.Record, error)
	FindRecordsByFilter(collectionNameOrId any, filter string, sort string, limit int, offset int, params ...dbx.Params) ([]*core.Record, error)
	FindRecordById(collectionNameOrId any, recordId string, optFilters ...func(*dbx.SelectQuery) error) (*core.Record, error)
	ExpandRecords(records []*core.Record, expands []string, optFetchFunc core.ExpandFetchFunc) map[string]error
	Save(model core.Model) error
	Delete(model core.Model) error
}

type AccountingService struct {
	app PocketBaseApp
}

func NewAccountingService(app PocketBaseApp) *AccountingService {
	return &AccountingService{app: app}
}

func (s *AccountingService) CreateCategory(input shared.CreateCategoryInput) (*shared.Category, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("category name is required")
	}
	if !isValidTransactionType(input.Type) {
		return nil, errors.New("invalid category type")
	}

	collection, err := s.app.FindCollectionByNameOrId("categories")
	if err != nil {
		return nil, fmt.Errorf("load categories collection: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("name", name)
	record.Set("type", string(input.Type))

	if err := s.app.Save(record); err != nil {
		return nil, fmt.Errorf("save category: %w", err)
	}

	return toCategory(record), nil
}

func (s *AccountingService) GetCategories() ([]shared.Category, error) {
	records, err := s.app.FindRecordsByFilter("categories", "", "type,name", 200, 0)
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}

	result := make([]shared.Category, 0, len(records))
	for _, r := range records {
		result = append(result, *toCategory(r))
	}

	return result, nil
}

func (s *AccountingService) UpdateCategory(input shared.UpdateCategoryInput) (*shared.Category, error) {
	id := strings.TrimSpace(input.ID)
	name := strings.TrimSpace(input.Name)
	if id == "" {
		return nil, errors.New("category id is required")
	}
	if name == "" {
		return nil, errors.New("category name is required")
	}
	if !isValidTransactionType(input.Type) {
		return nil, errors.New("invalid category type")
	}

	if err := s.ensureCategoryWithoutTransactions(id); err != nil {
		return nil, err
	}

	record, err := s.app.FindRecordById("categories", id)
	if err != nil {
		return nil, fmt.Errorf("find category: %w", err)
	}

	record.Set("name", name)
	record.Set("type", string(input.Type))
	if err := s.app.Save(record); err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}

	return toCategory(record), nil
}

func (s *AccountingService) DeleteCategory(categoryID string) error {
	id := strings.TrimSpace(categoryID)
	if id == "" {
		return errors.New("category id is required")
	}

	if err := s.ensureCategoryWithoutTransactions(id); err != nil {
		return err
	}

	record, err := s.app.FindRecordById("categories", id)
	if err != nil {
		return fmt.Errorf("find category: %w", err)
	}
	if err := s.app.Delete(record); err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	return nil
}

func (s *AccountingService) CreateTransaction(input shared.CreateTransactionInput) (*shared.Transaction, error) {
	input.Description = strings.TrimSpace(input.Description)
	if !isValidTransactionType(input.Type) {
		return nil, errors.New("invalid transaction type")
	}
	if input.Amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}
	if strings.TrimSpace(input.CategoryID) == "" {
		return nil, errors.New("category is required")
	}

	categoriesCollection, err := s.app.FindCollectionByNameOrId("categories")
	if err != nil {
		return nil, fmt.Errorf("load categories collection: %w", err)
	}
	categoryRecord, err := s.app.FindRecordsByFilter(categoriesCollection.Name, "id = {:id}", "", 1, 0, dbx.Params{"id": input.CategoryID})
	if err != nil {
		return nil, fmt.Errorf("check category existence: %w", err)
	}
	if len(categoryRecord) == 0 {
		return nil, errors.New("category not found")
	}
	if shared.TransactionType(categoryRecord[0].GetString("type")) != input.Type {
		return nil, errors.New("transaction type and category type must match")
	}

	transactionsCollection, err := s.app.FindCollectionByNameOrId("transactions")
	if err != nil {
		return nil, fmt.Errorf("load transactions collection: %w", err)
	}

	record := core.NewRecord(transactionsCollection)
	record.Set("type", string(input.Type))
	record.Set("amount", input.Amount)
	record.Set("description", input.Description)
	record.Set("category_id", input.CategoryID)

	if err := s.app.Save(record); err != nil {
		return nil, fmt.Errorf("save transaction: %w", err)
	}

	return &shared.Transaction{
		ID:          record.Id,
		Type:        shared.TransactionType(record.GetString("type")),
		Amount:      record.GetFloat("amount"),
		Description: record.GetString("description"),
		CategoryID:  record.GetString("category_id"),
		Category:    categoryRecord[0].GetString("name"),
		CreatedAt:   parseRecordTime(record, "created_at"),
	}, nil
}

func (s *AccountingService) GetTransactions() ([]shared.Transaction, error) {
	return s.GetTransactionsFiltered(shared.TransactionFilterInput{})
}

func (s *AccountingService) GetTransactionsFiltered(input shared.TransactionFilterInput) ([]shared.Transaction, error) {
	records, err := s.queryTransactionRecords(input)
	if err != nil {
		return nil, err
	}

	result := make([]shared.Transaction, 0, len(records))
	for _, r := range records {
		categoryName := ""
		if expanded := r.ExpandedOne("category_id"); expanded != nil {
			categoryName = expanded.GetString("name")
		}

		result = append(result, shared.Transaction{
			ID:          r.Id,
			Type:        shared.TransactionType(r.GetString("type")),
			Amount:      r.GetFloat("amount"),
			Description: r.GetString("description"),
			CategoryID:  r.GetString("category_id"),
			Category:    categoryName,
			CreatedAt:   parseRecordTime(r, "created_at"),
		})
	}

	return result, nil
}

func (s *AccountingService) ExportTransactionsCSV(input shared.TransactionFilterInput) (string, error) {
	transactions, err := s.GetTransactionsFiltered(input)
	if err != nil {
		return "", err
	}

	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)

	rows := [][]string{{"Fecha", "Tipo", "Categoria", "Descripcion", "Monto"}}
	for _, tx := range transactions {
		rows = append(rows, []string{
			tx.CreatedAt.Format("2006-01-02 15:04:05"),
			string(tx.Type),
			tx.Category,
			tx.Description,
			strconv.FormatFloat(tx.Amount, 'f', 2, 64),
		})
	}

	if err := writer.WriteAll(rows); err != nil {
		return "", fmt.Errorf("build csv content: %w", err)
	}

	return buffer.String(), nil
}

func (s *AccountingService) GetBalance() (*shared.Balance, error) {
	records, err := s.app.FindAllRecords("transactions")
	if err != nil {
		return nil, fmt.Errorf("query transactions for balance: %w", err)
	}

	balance := &shared.Balance{}
	for _, r := range records {
		amount := r.GetFloat("amount")
		typeValue := shared.TransactionType(r.GetString("type"))
		if typeValue == shared.TransactionTypeIncome {
			balance.Income += amount
		} else if typeValue == shared.TransactionTypeExpense {
			balance.Expense += amount
		}
	}
	balance.Total = balance.Income - balance.Expense

	return balance, nil
}

func (s *AccountingService) GetDashboardSummary() (*shared.DashboardSummary, error) {
	balance, err := s.GetBalance()
	if err != nil {
		return nil, err
	}

	transactions, err := s.GetTransactions()
	if err != nil {
		return nil, err
	}

	categoryMap := map[string]*shared.AccountingCategorySummary{}

	for _, tx := range transactions {
		key := strings.TrimSpace(tx.CategoryID)
		if key == "" {
			key = "uncategorized"
		}

		item := categoryMap[key]
		if item == nil {
			name := strings.TrimSpace(tx.Category)
			if name == "" {
				name = "Sin categoria"
			}
			item = &shared.AccountingCategorySummary{
				CategoryID: tx.CategoryID,
				Name:       name,
				Type:       tx.Type,
			}
			categoryMap[key] = item
		}

		if tx.Type == shared.TransactionTypeIncome {
			item.Income += tx.Amount
		} else if tx.Type == shared.TransactionTypeExpense {
			item.Expense += tx.Amount
		}
		item.Net = item.Income - item.Expense
		item.Transactions++
		if tx.CreatedAt.After(item.LastActivity) {
			item.LastActivity = tx.CreatedAt
		}
	}

	breakdown := make([]shared.AccountingCategorySummary, 0, len(categoryMap))
	for _, item := range categoryMap {
		breakdown = append(breakdown, *item)
	}

	sort.SliceStable(breakdown, func(i, j int) bool {
		iTotal := breakdown[i].Income + breakdown[i].Expense
		jTotal := breakdown[j].Income + breakdown[j].Expense
		if iTotal == jTotal {
			return strings.ToLower(breakdown[i].Name) < strings.ToLower(breakdown[j].Name)
		}
		return iTotal > jTotal
	})

	return &shared.DashboardSummary{
		Balance:           *balance,
		CategoryBreakdown: breakdown,
	}, nil
}

func (s *AccountingService) GetCategoryRanking(input shared.CategoryRankingInput) (*shared.CategoryRankingResult, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 5
	}
	metric := input.Metric
	if metric == "" {
		metric = shared.RankingByExpense
	}
	if !isValidRankingMetric(metric) {
		return nil, errors.New("invalid ranking metric")
	}

	since, err := s.rankingPeriodStart()
	if err != nil {
		return nil, err
	}
	now := time.Now()

	filter := "created_at >= {:since}"
	params := dbx.Params{"since": since.Format(time.RFC3339)}
	if metric == shared.RankingByIncome || metric == shared.RankingByExpense {
		filter += " && type = {:type}"
		params["type"] = string(metric)
	}

	records, err := s.app.FindRecordsByFilter("transactions", filter, "-created_at", 2000, 0, params)
	if err != nil {
		return nil, fmt.Errorf("query transactions for ranking: %w", err)
	}

	if errs := s.app.ExpandRecords(records, []string{"category_id"}, nil); len(errs) > 0 {
		return nil, fmt.Errorf("expand category relation: %v", errs)
	}

	type agg struct {
		categoryID   string
		name         string
		income       float64
		expense      float64
		transactions int
	}

	index := map[string]*agg{}
	order := make([]string, 0, len(records))
	var totalIncome, totalExpense float64
	for _, r := range records {
		categoryID := r.GetString("category_id")
		entry := index[categoryID]
		if entry == nil {
			name := ""
			if expanded := r.ExpandedOne("category_id"); expanded != nil {
				name = expanded.GetString("name")
			}
			if name == "" {
				name = "Sin categoria"
			}
			entry = &agg{categoryID: categoryID, name: name}
			index[categoryID] = entry
			order = append(order, categoryID)
		}

		amount := r.GetFloat("amount")
		entry.transactions++
		if shared.TransactionType(r.GetString("type")) == shared.TransactionTypeIncome {
			entry.income += amount
			totalIncome += amount
		} else {
			entry.expense += amount
			totalExpense += amount
		}
	}

	items := make([]shared.CategoryRankingItem, 0, len(order))
	for _, id := range order {
		a := index[id]
		items = append(items, shared.CategoryRankingItem{
			CategoryID:   a.categoryID,
			Name:         a.name,
			Income:       a.income,
			Expense:      a.expense,
			Transactions: a.transactions,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		vi := rankingValue(items[i], metric)
		vj := rankingValue(items[j], metric)
		if vi == vj {
			return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		}
		return vi > vj
	})

	if len(items) > limit {
		items = items[:limit]
	}

	return &shared.CategoryRankingResult{
		Items:        items,
		StartDate:    since.Format("2006-01-02"),
		EndDate:      now.Format("2006-01-02"),
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
	}, nil
}

func (s *AccountingService) rankingPeriodStart() (time.Time, error) {
	settings, err := s.app.FindRecordsByFilter("app_settings", "", "", 1, 0)
	if err != nil {
		return time.Time{}, fmt.Errorf("query app settings: %w", err)
	}

	days := 0
	if len(settings) > 0 {
		days = settings[0].GetInt("ranking_days")
	}

	now := time.Now()
	if days > 0 {
		return now.AddDate(0, 0, -days), nil
	}

	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), nil
}

func (s *AccountingService) GetAppSettings() (*shared.AppSettings, error) {
	records, err := s.app.FindRecordsByFilter("app_settings", "", "", 1, 0)
	if err != nil {
		return nil, fmt.Errorf("query app settings: %w", err)
	}
	if len(records) == 0 {
		return &shared.AppSettings{RankingDays: 0, WeekStartDay: defaultWeekStartDay, WeeklyBudget: 0}, nil
	}

	return &shared.AppSettings{
		RankingDays:  records[0].GetInt("ranking_days"),
		WeekStartDay: normalizeWeekStartDay(records[0].GetInt("week_start_day")),
		WeeklyBudget: records[0].GetFloat("weekly_budget"),
	}, nil
}

func (s *AccountingService) SaveAppSettings(input shared.AppSettings) (*shared.AppSettings, error) {
	days := input.RankingDays
	if days < 0 {
		days = 0
	}
	weekStartDay := normalizeWeekStartDay(input.WeekStartDay)
	weeklyBudget := input.WeeklyBudget
	if weeklyBudget < 0 {
		weeklyBudget = 0
	}

	collection, err := s.app.FindCollectionByNameOrId("app_settings")
	if err != nil {
		return nil, fmt.Errorf("load app settings collection: %w", err)
	}

	records, err := s.app.FindRecordsByFilter("app_settings", "", "", 1, 0)
	if err != nil {
		return nil, fmt.Errorf("query app settings: %w", err)
	}

	var record *core.Record
	if len(records) == 0 {
		record = core.NewRecord(collection)
	} else {
		record = records[0]
	}

	record.Set("ranking_days", days)
	record.Set("week_start_day", weekStartDay)
	record.Set("weekly_budget", weeklyBudget)
	if err := s.app.Save(record); err != nil {
		return nil, fmt.Errorf("save app settings: %w", err)
	}

	return &shared.AppSettings{RankingDays: days, WeekStartDay: weekStartDay, WeeklyBudget: weeklyBudget}, nil
}

func (s *AccountingService) GetBudgets() ([]shared.Budget, error) {
	records, err := s.app.FindRecordsByFilter("budgets", "", "", 200, 0)
	if err != nil {
		return nil, fmt.Errorf("query budgets: %w", err)
	}

	if errs := s.app.ExpandRecords(records, []string{"category_id"}, nil); len(errs) > 0 {
		return nil, fmt.Errorf("expand budget category: %v", errs)
	}

	result := make([]shared.Budget, 0, len(records))
	for _, r := range records {
		categoryName := ""
		if expanded := r.ExpandedOne("category_id"); expanded != nil {
			categoryName = expanded.GetString("name")
		}

		result = append(result, shared.Budget{
			ID:           r.Id,
			CategoryID:   r.GetString("category_id"),
			CategoryName: categoryName,
			WeeklyAmount: r.GetFloat("weekly_amount"),
			Active:       r.GetBool("active"),
		})
	}

	return result, nil
}

func (s *AccountingService) SaveBudget(input shared.SaveBudgetInput) (*shared.Budget, error) {
	categoryID := strings.TrimSpace(input.CategoryID)
	if categoryID == "" {
		return nil, errors.New("category is required")
	}
	if input.WeeklyAmount < 0 {
		return nil, errors.New("weekly amount cannot be negative")
	}

	category, err := s.app.FindRecordById("categories", categoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}
	if shared.TransactionType(category.GetString("type")) != shared.TransactionTypeExpense {
		return nil, errors.New("budget can only be set for expense categories")
	}

	collection, err := s.app.FindCollectionByNameOrId("budgets")
	if err != nil {
		return nil, fmt.Errorf("load budgets collection: %w", err)
	}

	existing, err := s.app.FindRecordsByFilter("budgets", "category_id = {:category_id}", "", 1, 0, dbx.Params{"category_id": categoryID})
	if err != nil {
		return nil, fmt.Errorf("query existing budget: %w", err)
	}

	var record *core.Record
	if len(existing) > 0 {
		record = existing[0]
	} else {
		record = core.NewRecord(collection)
		record.Set("category_id", categoryID)
		record.Set("active", true)
	}
	record.Set("weekly_amount", input.WeeklyAmount)
	if err := s.app.Save(record); err != nil {
		return nil, fmt.Errorf("save budget: %w", err)
	}

	return &shared.Budget{
		ID:           record.Id,
		CategoryID:   record.GetString("category_id"),
		CategoryName: category.GetString("name"),
		WeeklyAmount: record.GetFloat("weekly_amount"),
		Active:       record.GetBool("active"),
	}, nil
}

func (s *AccountingService) DeleteBudget(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("budget id is required")
	}

	record, err := s.app.FindRecordById("budgets", id)
	if err != nil {
		return fmt.Errorf("find budget: %w", err)
	}

	if err := s.app.Delete(record); err != nil {
		return fmt.Errorf("delete budget: %w", err)
	}

	return nil
}

func (s *AccountingService) GetWeeklySummary(weekOffset int) (*shared.WeeklySummary, error) {
	startDay := s.weekStartDay()
	now := time.Now()
	weekStart := startOfWeek(now, startDay).AddDate(0, 0, 7*weekOffset)
	weekEnd := weekStart.AddDate(0, 0, 7)

	spent, err := s.expensesByCategory(weekStart, weekEnd)
	if err != nil {
		return nil, err
	}

	prevSpent, err := s.expensesByCategory(weekStart.AddDate(0, 0, -7), weekStart)
	if err != nil {
		return nil, err
	}

	budgets, err := s.activeBudgetMap()
	if err != nil {
		return nil, err
	}

	ids := make(map[string]bool, len(spent)+len(prevSpent)+len(budgets))
	for id := range budgets {
		ids[id] = true
	}
	for id := range spent {
		ids[id] = true
	}
	for id := range prevSpent {
		ids[id] = true
	}

	categories := make([]shared.WeeklyCategorySummary, 0, len(ids))
	var totalSpent, totalBudget float64
	for id := range ids {
		name := ""
		cur := 0.0
		prev := 0.0
		budget := 0.0

		if b, ok := budgets[id]; ok {
			budget = b.WeeklyAmount
			if name == "" {
				name = b.CategoryName
			}
		}
		if c, ok := spent[id]; ok {
			cur = c.spent
			if name == "" {
				name = c.name
			}
		}
		if c, ok := prevSpent[id]; ok {
			prev = c.spent
			if name == "" {
				name = c.name
			}
		}
		if name == "" {
			name = "Sin categoría"
		}

		percent := 0.0
		if budget > 0 {
			percent = cur / budget * 100
		}

		categories = append(categories, shared.WeeklyCategorySummary{
			CategoryID:   id,
			CategoryName: name,
			Spent:        cur,
			Budget:       budget,
			PercentUsed:  percent,
			Remaining:    budget - cur,
			DeltaVsPrev:  cur - prev,
		})

		totalSpent += cur
		totalBudget += budget
	}

	sort.SliceStable(categories, func(i, j int) bool {
		if categories[i].Spent == categories[j].Spent {
			return strings.ToLower(categories[i].CategoryName) < strings.ToLower(categories[j].CategoryName)
		}
		return categories[i].Spent > categories[j].Spent
	})

	var prevTotal float64
	for _, c := range prevSpent {
		prevTotal += c.spent
	}

	totalPercent := 0.0
	if totalBudget > 0 {
		totalPercent = totalSpent / totalBudget * 100
	}

	globalBudget := s.globalWeeklyBudget()
	globalPercent := 0.0
	if globalBudget > 0 {
		globalPercent = totalSpent / globalBudget * 100
	}

	balance, err := s.GetBalance()
	if err != nil {
		return nil, err
	}

	return &shared.WeeklySummary{
		StartDate:        weekStart.Format("2006-01-02"),
		EndDate:          weekEnd.AddDate(0, 0, -1).Format("2006-01-02"),
		TotalSpent:       totalSpent,
		TotalBudget:      totalBudget,
		TotalPercent:     totalPercent,
		TotalRemaining:   totalBudget - totalSpent,
		DeltaVsPrev:      totalSpent - prevTotal,
		GlobalBudget:     globalBudget,
		GlobalPercent:    globalPercent,
		GlobalRemaining:  globalBudget - totalSpent,
		AvailableBalance: balance.Total,
		Categories:       categories,
	}, nil
}

type categoryExpense struct {
	name  string
	spent float64
}

func (s *AccountingService) expensesByCategory(start, end time.Time) (map[string]*categoryExpense, error) {
	records, err := s.app.FindRecordsByFilter("transactions",
		"type = {:type} && created_at >= {:start} && created_at < {:end}",
		"", 2000, 0, dbx.Params{
			"type":  string(shared.TransactionTypeExpense),
			"start": start.Format(time.RFC3339),
			"end":   end.Format(time.RFC3339),
		})
	if err != nil {
		return nil, fmt.Errorf("query weekly expenses: %w", err)
	}

	if errs := s.app.ExpandRecords(records, []string{"category_id"}, nil); len(errs) > 0 {
		return nil, fmt.Errorf("expand expense category: %v", errs)
	}

	result := make(map[string]*categoryExpense)
	for _, r := range records {
		categoryID := r.GetString("category_id")
		if categoryID == "" {
			continue
		}

		entry := result[categoryID]
		if entry == nil {
			name := ""
			if expanded := r.ExpandedOne("category_id"); expanded != nil {
				name = expanded.GetString("name")
			}
			if name == "" {
				name = "Sin categoría"
			}
			entry = &categoryExpense{name: name}
			result[categoryID] = entry
		}

		entry.spent += r.GetFloat("amount")
	}

	return result, nil
}

func (s *AccountingService) activeBudgetMap() (map[string]*shared.Budget, error) {
	records, err := s.app.FindRecordsByFilter("budgets", "active = true", "", 200, 0)
	if err != nil {
		return nil, fmt.Errorf("query active budgets: %w", err)
	}

	if errs := s.app.ExpandRecords(records, []string{"category_id"}, nil); len(errs) > 0 {
		return nil, fmt.Errorf("expand active budget category: %v", errs)
	}

	result := make(map[string]*shared.Budget, len(records))
	for _, r := range records {
		categoryName := ""
		if expanded := r.ExpandedOne("category_id"); expanded != nil {
			categoryName = expanded.GetString("name")
		}

		result[r.GetString("category_id")] = &shared.Budget{
			ID:           r.Id,
			CategoryID:   r.GetString("category_id"),
			CategoryName: categoryName,
			WeeklyAmount: r.GetFloat("weekly_amount"),
			Active:       r.GetBool("active"),
		}
	}

	return result, nil
}

func (s *AccountingService) weekStartDay() int {
	records, err := s.app.FindRecordsByFilter("app_settings", "", "", 1, 0)
	if err != nil || len(records) == 0 {
		return defaultWeekStartDay
	}
	return normalizeWeekStartDay(records[0].GetInt("week_start_day"))
}

func (s *AccountingService) globalWeeklyBudget() float64 {
	records, err := s.app.FindRecordsByFilter("app_settings", "", "", 1, 0)
	if err != nil || len(records) == 0 {
		return 0
	}
	budget := records[0].GetFloat("weekly_budget")
	if budget < 0 {
		return 0
	}
	return budget
}

const defaultWeekStartDay = 1

func normalizeWeekStartDay(value int) int {
	if value < 1 || value > 7 {
		return defaultWeekStartDay
	}
	return value
}

func startOfWeek(reference time.Time, weekStartDay int) time.Time {
	current := time.Date(reference.Year(), reference.Month(), reference.Day(), 0, 0, 0, 0, reference.Location())
	target := weekStartDay % 7
	daysSince := (int(current.Weekday()) - target + 7) % 7
	return current.AddDate(0, 0, -daysSince)
}

func rankingValue(item shared.CategoryRankingItem, metric shared.CategoryRankingMetric) float64 {
	switch metric {
	case shared.RankingByIncome:
		return item.Income
	case shared.RankingByTransactions:
		return float64(item.Transactions)
	default:
		return item.Expense
	}
}

func isValidRankingMetric(v shared.CategoryRankingMetric) bool {
	return v == shared.RankingByExpense || v == shared.RankingByIncome || v == shared.RankingByTransactions
}

func (s *AccountingService) GetAIProviders() ([]shared.AIProvider, error) {
	records, err := s.app.FindRecordsByFilter("ai_providers", "", "name", 200, 0)
	if err != nil {
		return nil, fmt.Errorf("query ai providers: %w", err)
	}

	result := make([]shared.AIProvider, 0, len(records))
	for _, r := range records {
		result = append(result, shared.AIProvider{
			ID:       r.Id,
			Name:     r.GetString("name"),
			BaseURL:  r.GetString("base_url"),
			ChatPath: r.GetString("chat_path"),
			APIKey:   r.GetString("api_key"),
			Active:   r.GetBool("active"),
		})
	}
	return result, nil
}

func (s *AccountingService) CreateAIProvider(input shared.CreateAIProviderInput) (*shared.AIProvider, error) {
	name := strings.TrimSpace(input.Name)
	baseURL := strings.TrimSpace(input.BaseURL)
	chatPath := strings.TrimSpace(input.ChatPath)
	if name == "" || baseURL == "" || chatPath == "" {
		return nil, errors.New("name, base_url and chat_path are required")
	}

	collection, err := s.app.FindCollectionByNameOrId("ai_providers")
	if err != nil {
		return nil, fmt.Errorf("load ai providers collection: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("name", name)
	record.Set("base_url", baseURL)
	record.Set("chat_path", chatPath)
	record.Set("active", true)

	if err := s.app.Save(record); err != nil {
		return nil, fmt.Errorf("save ai provider: %w", err)
	}

	return &shared.AIProvider{
		ID:       record.Id,
		Name:     record.GetString("name"),
		BaseURL:  record.GetString("base_url"),
		ChatPath: record.GetString("chat_path"),
		APIKey:   record.GetString("api_key"),
		Active:   record.GetBool("active"),
	}, nil
}

func (s *AccountingService) GetAIModels(providerID string) ([]shared.AIModel, error) {
	filter := ""
	params := dbx.Params{}
	if strings.TrimSpace(providerID) != "" {
		filter = "provider_id = {:provider_id}"
		params["provider_id"] = strings.TrimSpace(providerID)
	}

	var records []*core.Record
	var err error
	if len(params) == 0 {
		records, err = s.app.FindRecordsByFilter("ai_models", filter, "name", 200, 0)
	} else {
		records, err = s.app.FindRecordsByFilter("ai_models", filter, "name", 200, 0, params)
	}
	if err != nil {
		return nil, fmt.Errorf("query ai models: %w", err)
	}

	result := make([]shared.AIModel, 0, len(records))
	for _, r := range records {
		result = append(result, shared.AIModel{
			ID:         r.Id,
			ProviderID: r.GetString("provider_id"),
			Name:       r.GetString("name"),
			Active:     r.GetBool("active"),
		})
	}
	return result, nil
}

func (s *AccountingService) CreateAIModel(input shared.CreateAIModelInput) (*shared.AIModel, error) {
	providerID := strings.TrimSpace(input.ProviderID)
	name := strings.TrimSpace(input.Name)
	if providerID == "" || name == "" {
		return nil, errors.New("provider_id and model name are required")
	}

	if _, err := s.app.FindRecordById("ai_providers", providerID); err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	collection, err := s.app.FindCollectionByNameOrId("ai_models")
	if err != nil {
		return nil, fmt.Errorf("load ai models collection: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("provider_id", providerID)
	record.Set("name", name)
	record.Set("active", true)
	if err := s.app.Save(record); err != nil {
		return nil, fmt.Errorf("save ai model: %w", err)
	}

	return &shared.AIModel{
		ID:         record.Id,
		ProviderID: record.GetString("provider_id"),
		Name:       record.GetString("name"),
		Active:     record.GetBool("active"),
	}, nil
}

func (s *AccountingService) GetAIConfiguration() (*shared.AIConfiguration, error) {
	settings, err := s.app.FindRecordsByFilter("ai_settings", "", "", 1, 0)
	if err != nil {
		return nil, fmt.Errorf("query ai settings: %w", err)
	}
	if len(settings) == 0 {
		return &shared.AIConfiguration{}, nil
	}

	return &shared.AIConfiguration{
		ProviderID: settings[0].GetString("provider_id"),
		ModelID:    settings[0].GetString("model_id"),
	}, nil
}

func (s *AccountingService) SaveAIConfiguration(input shared.SaveAIConfigurationInput) (*shared.AIConfiguration, error) {
	providerID := strings.TrimSpace(input.ProviderID)
	modelID := strings.TrimSpace(input.ModelID)
	if providerID == "" || modelID == "" {
		return nil, errors.New("provider and model are required")
	}

	provider, err := s.app.FindRecordById("ai_providers", providerID)
	if err != nil {
		return nil, fmt.Errorf("find provider: %w", err)
	}

	if _, err := s.app.FindRecordById("ai_models", modelID); err != nil {
		return nil, fmt.Errorf("find model: %w", err)
	}

	apiKey := strings.TrimSpace(input.APIKey)
	if apiKey != "" {
		provider.Set("api_key", apiKey)
		if err := s.app.Save(provider); err != nil {
			return nil, fmt.Errorf("save provider api key: %w", err)
		}
	}

	collection, err := s.app.FindCollectionByNameOrId("ai_settings")
	if err != nil {
		return nil, fmt.Errorf("load ai settings collection: %w", err)
	}

	settings, err := s.app.FindRecordsByFilter("ai_settings", "", "", 1, 0)
	if err != nil {
		return nil, fmt.Errorf("query ai settings: %w", err)
	}

	var record *core.Record
	if len(settings) == 0 {
		record = core.NewRecord(collection)
	} else {
		record = settings[0]
	}
	record.Set("provider_id", providerID)
	record.Set("model_id", modelID)
	if err := s.app.Save(record); err != nil {
		return nil, fmt.Errorf("save ai settings: %w", err)
	}

	return &shared.AIConfiguration{ProviderID: providerID, ModelID: modelID}, nil
}

func (s *AccountingService) AnalyzeDashboardWithAI(input shared.AnalyzeDashboardInput) (*shared.AIAnalysisRecord, error) {
	config, err := s.GetAIConfiguration()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(config.ProviderID) == "" || strings.TrimSpace(config.ModelID) == "" {
		return nil, errors.New("configure ai provider and model first")
	}

	provider, err := s.app.FindRecordById("ai_providers", config.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("find configured provider: %w", err)
	}
	model, err := s.app.FindRecordById("ai_models", config.ModelID)
	if err != nil {
		return nil, fmt.Errorf("find configured model: %w", err)
	}

	apiKey := strings.TrimSpace(provider.GetString("api_key"))
	if apiKey == "" {
		return nil, errors.New("api key is missing for selected provider")
	}

	summary, err := s.GetDashboardSummary()
	if err != nil {
		return nil, err
	}

	payload := map[string]any{
		"model": model.GetString("name"),
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Eres un analista financiero para contabilidad personal. Responde en español con recomendaciones claras y accionables.",
			},
			{
				"role": "user",
				"content": fmt.Sprintf(
					"Analiza este dashboard: balance total %.2f, ingresos %.2f, egresos %.2f, categorias %s. Pregunta adicional: %s",
					summary.Balance.Total,
					summary.Balance.Income,
					summary.Balance.Expense,
					summarizeCategoryBreakdown(summary.CategoryBreakdown),
					strings.TrimSpace(input.Question),
				),
			},
		},
		"temperature": 0.4,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal ai payload: %w", err)
	}

	endpoint, err := buildProviderEndpoint(provider.GetString("base_url"), provider.GetString("chat_path"))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create ai request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 35 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ai provider request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read ai response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ai provider error: %s", strings.TrimSpace(string(raw)))
	}

	var aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &aiResp); err != nil {
		return nil, fmt.Errorf("decode ai response: %w", err)
	}
	if len(aiResp.Choices) == 0 || strings.TrimSpace(aiResp.Choices[0].Message.Content) == "" {
		return nil, errors.New("ai response was empty")
	}

	analysis := strings.TrimSpace(aiResp.Choices[0].Message.Content)

	analysesCollection, err := s.app.FindCollectionByNameOrId("ai_analyses")
	if err != nil {
		return nil, fmt.Errorf("load ai analyses collection: %w", err)
	}

	record := core.NewRecord(analysesCollection)
	record.Set("question", strings.TrimSpace(input.Question))
	record.Set("provider", provider.GetString("name"))
	record.Set("model", model.GetString("name"))
	record.Set("analysis", analysis)
	if err := s.app.Save(record); err != nil {
		return nil, fmt.Errorf("save ai analysis: %w", err)
	}

	return &shared.AIAnalysisRecord{
		ID:        record.Id,
		Provider:  provider.GetString("name"),
		Model:     model.GetString("name"),
		Question:  strings.TrimSpace(input.Question),
		Analysis:  analysis,
		CreatedAt: parseRecordTime(record, "created_at"),
	}, nil
}

func (s *AccountingService) GetAIAnalyses() ([]shared.AIAnalysisRecord, error) {
	records, err := s.app.FindRecordsByFilter("ai_analyses", "", "-created_at", 200, 0)
	if err != nil {
		return nil, fmt.Errorf("query ai analyses: %w", err)
	}

	result := make([]shared.AIAnalysisRecord, 0, len(records))
	for _, r := range records {
		result = append(result, shared.AIAnalysisRecord{
			ID:        r.Id,
			Provider:  r.GetString("provider"),
			Model:     r.GetString("model"),
			Question:  r.GetString("question"),
			Analysis:  r.GetString("analysis"),
			CreatedAt: parseRecordTime(r, "created_at"),
		})
	}

	return result, nil
}

func (s *AccountingService) DeleteAIAnalysis(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("analysis id is required")
	}

	record, err := s.app.FindRecordById("ai_analyses", id)
	if err != nil {
		return fmt.Errorf("find ai analysis: %w", err)
	}

	if err := s.app.Delete(record); err != nil {
		return fmt.Errorf("delete ai analysis: %w", err)
	}

	return nil
}

func (s *AccountingService) CreateBillingCategory(input shared.CreateBillingCategoryInput) (*shared.BillingCategory, error) {
	name := strings.TrimSpace(input.Name)
	unitLabel := strings.TrimSpace(input.UnitLabel)
	if name == "" {
		return nil, errors.New("billing category name is required")
	}
	if !isValidBillingCalcType(input.CalcType) {
		return nil, errors.New("invalid billing calc type")
	}
	if input.DefaultRate < 0 {
		return nil, errors.New("default rate cannot be negative")
	}
	if unitLabel == "" {
		if input.CalcType == shared.BillingCalcTypeHourly {
			unitLabel = "hora"
		} else {
			unitLabel = "unidad"
		}
	}
	if !isValidTransactionType(input.TransactionType) {
		return nil, errors.New("invalid transaction type")
	}

	collection, err := s.app.FindCollectionByNameOrId("billing_categories")
	if err != nil {
		return nil, fmt.Errorf("load billing categories collection: %w", err)
	}

	record := core.NewRecord(collection)
	record.Set("name", name)
	record.Set("calc_type", string(input.CalcType))
	record.Set("default_rate", input.DefaultRate)
	record.Set("unit_label", unitLabel)
	record.Set("transaction_type", string(input.TransactionType))

	if err := s.app.Save(record); err != nil {
		return nil, fmt.Errorf("save billing category: %w", err)
	}

	return toBillingCategory(record), nil
}

func (s *AccountingService) GetBillingCategories() ([]shared.BillingCategory, error) {
	records, err := s.app.FindRecordsByFilter("billing_categories", "", "name", 200, 0)
	if err != nil {
		return nil, fmt.Errorf("query billing categories: %w", err)
	}

	result := make([]shared.BillingCategory, 0, len(records))
	for _, r := range records {
		result = append(result, *toBillingCategory(r))
	}

	return result, nil
}

func (s *AccountingService) UpdateBillingCategory(input shared.UpdateBillingCategoryInput) (*shared.BillingCategory, error) {
	id := strings.TrimSpace(input.ID)
	name := strings.TrimSpace(input.Name)
	unitLabel := strings.TrimSpace(input.UnitLabel)
	if id == "" {
		return nil, errors.New("billing category id is required")
	}
	if name == "" {
		return nil, errors.New("billing category name is required")
	}
	if !isValidBillingCalcType(input.CalcType) {
		return nil, errors.New("invalid billing calc type")
	}
	if input.DefaultRate < 0 {
		return nil, errors.New("default rate cannot be negative")
	}
	if unitLabel == "" {
		if input.CalcType == shared.BillingCalcTypeHourly {
			unitLabel = "hora"
		} else {
			unitLabel = "unidad"
		}
	}
	if !isValidTransactionType(input.TransactionType) {
		return nil, errors.New("invalid transaction type")
	}

	record, err := s.app.FindRecordById("billing_categories", id)
	if err != nil {
		return nil, fmt.Errorf("find billing category: %w", err)
	}

	record.Set("name", name)
	record.Set("calc_type", string(input.CalcType))
	record.Set("default_rate", input.DefaultRate)
	record.Set("unit_label", unitLabel)
	record.Set("transaction_type", string(input.TransactionType))

	if err := s.app.Save(record); err != nil {
		return nil, fmt.Errorf("update billing category: %w", err)
	}

	return toBillingCategory(record), nil
}

func (s *AccountingService) DeleteBillingCategory(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("billing category id is required")
	}

	record, err := s.app.FindRecordById("billing_categories", id)
	if err != nil {
		return fmt.Errorf("find billing category: %w", err)
	}

	if err := s.app.Delete(record); err != nil {
		return fmt.Errorf("delete billing category: %w", err)
	}

	return nil
}

func (s *AccountingService) CalculateQuote(input shared.CalculateQuoteInput) (*shared.QuoteResult, error) {
	id := strings.TrimSpace(input.BillingCategoryID)
	if id == "" {
		return nil, errors.New("billing category is required")
	}
	if input.Quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	record, err := s.app.FindRecordById("billing_categories", id)
	if err != nil {
		return nil, fmt.Errorf("find billing category: %w", err)
	}

	rate := input.Rate
	if rate <= 0 {
		rate = record.GetFloat("default_rate")
	}
	if rate <= 0 {
		return nil, errors.New("rate must be greater than 0")
	}

	calcType := shared.BillingCalcType(record.GetString("calc_type"))
	unitLabel := strings.TrimSpace(record.GetString("unit_label"))
	if unitLabel == "" {
		if calcType == shared.BillingCalcTypeHourly {
			unitLabel = "hora"
		} else {
			unitLabel = "unidad"
		}
	}

	amount := input.Quantity * rate
	text := strings.TrimSpace(input.Description)
	if text == "" {
		text = fmt.Sprintf("%s: %.2f %s x %.2f", record.GetString("name"), input.Quantity, unitLabel, rate)
	}

	return &shared.QuoteResult{
		BillingCategoryID: record.Id,
		CategoryName:      record.GetString("name"),
		CalcType:          calcType,
		UnitLabel:         unitLabel,
		Quantity:          input.Quantity,
		Rate:              rate,
		Amount:            amount,
		Description:       text,
		SuggestedType:     shared.TransactionType(record.GetString("transaction_type")),
	}, nil
}

func (s *AccountingService) queryTransactionRecords(input shared.TransactionFilterInput) ([]*core.Record, error) {
	filter, params, err := buildTransactionFilter(input)
	if err != nil {
		return nil, err
	}

	records, err := s.app.FindRecordsByFilter("transactions", filter, "-created_at", 1000, 0, params)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}

	errs := s.app.ExpandRecords(records, []string{"category_id"}, nil)
	if len(errs) > 0 {
		return nil, fmt.Errorf("expand category relation: %v", errs)
	}

	return records, nil
}

func buildTransactionFilter(input shared.TransactionFilterInput) (string, dbx.Params, error) {
	parts := make([]string, 0, 4)
	params := dbx.Params{}

	if strings.TrimSpace(input.CategoryID) != "" {
		parts = append(parts, "category_id = {:category_id}")
		params["category_id"] = strings.TrimSpace(input.CategoryID)
	}

	if strings.TrimSpace(string(input.Type)) != "" {
		if !isValidTransactionType(input.Type) {
			return "", nil, errors.New("invalid transaction type filter")
		}
		parts = append(parts, "type = {:type}")
		params["type"] = string(input.Type)
	}

	if strings.TrimSpace(input.StartDate) != "" {
		start, err := parseFilterDate(input.StartDate)
		if err != nil {
			return "", nil, fmt.Errorf("invalid start date: %w", err)
		}
		parts = append(parts, "created_at >= {:start_date}")
		params["start_date"] = start.Format(time.RFC3339)
	}

	if strings.TrimSpace(input.EndDate) != "" {
		end, err := parseFilterDate(input.EndDate)
		if err != nil {
			return "", nil, fmt.Errorf("invalid end date: %w", err)
		}
		end = end.Add(24 * time.Hour)
		parts = append(parts, "created_at < {:end_date}")
		params["end_date"] = end.Format(time.RFC3339)
	}

	if params["start_date"] != nil && params["end_date"] != nil {
		start := params["start_date"].(string)
		end := params["end_date"].(string)
		if end <= start {
			return "", nil, errors.New("end date must be on or after start date")
		}
	}

	return strings.Join(parts, " && "), params, nil
}

func parseFilterDate(value string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, err
	}
	return date, nil
}

func summarizeCategoryBreakdown(items []shared.AccountingCategorySummary) string {
	if len(items) == 0 {
		return "sin movimientos por categoria"
	}

	parts := make([]string, 0, len(items))
	for _, item := range items {
		parts = append(parts, fmt.Sprintf("%s: ingresos %.2f, egresos %.2f, neto %.2f", item.Name, item.Income, item.Expense, item.Net))
	}
	return strings.Join(parts, " | ")
}

func buildProviderEndpoint(baseURL, chatPath string) (string, error) {
	baseURL = strings.TrimSpace(baseURL)
	chatPath = strings.TrimSpace(chatPath)
	if baseURL == "" || chatPath == "" {
		return "", errors.New("provider base_url and chat_path are required")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid provider base_url: %w", err)
	}
	if !strings.HasPrefix(chatPath, "/") {
		chatPath = "/" + chatPath
	}
	base.Path = strings.TrimSuffix(base.Path, "/") + chatPath
	return base.String(), nil
}

func (s *AccountingService) ensureCategoryWithoutTransactions(categoryID string) error {
	tx, err := s.app.FindRecordsByFilter("transactions", "category_id = {:category_id}", "", 1, 0, dbx.Params{"category_id": categoryID})
	if err != nil {
		return fmt.Errorf("check category movements: %w", err)
	}
	if len(tx) > 0 {
		return errors.New("category has transactions and cannot be modified or deleted")
	}
	return nil
}

func isValidTransactionType(v shared.TransactionType) bool {
	return v == shared.TransactionTypeIncome || v == shared.TransactionTypeExpense
}

func isValidBillingCalcType(v shared.BillingCalcType) bool {
	return v == shared.BillingCalcTypeHourly || v == shared.BillingCalcTypeUnit
}

func toCategory(record *core.Record) *shared.Category {
	return &shared.Category{
		ID:   record.Id,
		Name: record.GetString("name"),
		Type: shared.TransactionType(record.GetString("type")),
	}
}

func toBillingCategory(record *core.Record) *shared.BillingCategory {
	return &shared.BillingCategory{
		ID:              record.Id,
		Name:            record.GetString("name"),
		CalcType:        shared.BillingCalcType(record.GetString("calc_type")),
		DefaultRate:     record.GetFloat("default_rate"),
		UnitLabel:       record.GetString("unit_label"),
		TransactionType: shared.TransactionType(record.GetString("transaction_type")),
	}
}

func parseRecordTime(record *core.Record, field string) time.Time {
	value := strings.TrimSpace(record.GetString(field))
	if value == "" {
		return time.Time{}
	}

	layoutCandidates := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.000Z", "2006-01-02 15:04:05"}
	for _, layout := range layoutCandidates {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed
		}
	}

	return time.Time{}
}
