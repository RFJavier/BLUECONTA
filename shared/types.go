package shared

import "time"

type TransactionType string

type BillingCalcType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"

	BillingCalcTypeHourly BillingCalcType = "hourly"
	BillingCalcTypeUnit   BillingCalcType = "unit"
)

type Category struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Type TransactionType `json:"type"`
}

type Transaction struct {
	ID          string          `json:"id"`
	Type        TransactionType `json:"type"`
	Amount      float64         `json:"amount"`
	Description string          `json:"description"`
	CategoryID  string          `json:"category_id"`
	Category    string          `json:"category"`
	CreatedAt   time.Time       `json:"created_at"`
}

type TransactionFilterInput struct {
	StartDate  string          `json:"start_date"`
	EndDate    string          `json:"end_date"`
	CategoryID string          `json:"category_id"`
	Type       TransactionType `json:"type"`
}

type CreateCategoryInput struct {
	Name string          `json:"name"`
	Type TransactionType `json:"type"`
}

type UpdateCategoryInput struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Type TransactionType `json:"type"`
}

type CreateTransactionInput struct {
	Type        TransactionType `json:"type"`
	Amount      float64         `json:"amount"`
	Description string          `json:"description"`
	CategoryID  string          `json:"category_id"`
}

type Balance struct {
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Total   float64 `json:"total"`
}

type BillingCategory struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	CalcType        BillingCalcType `json:"calc_type"`
	DefaultRate     float64         `json:"default_rate"`
	UnitLabel       string          `json:"unit_label"`
	TransactionType TransactionType `json:"transaction_type"`
}

type CreateBillingCategoryInput struct {
	Name            string          `json:"name"`
	CalcType        BillingCalcType `json:"calc_type"`
	DefaultRate     float64         `json:"default_rate"`
	UnitLabel       string          `json:"unit_label"`
	TransactionType TransactionType `json:"transaction_type"`
}

type UpdateBillingCategoryInput struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	CalcType        BillingCalcType `json:"calc_type"`
	DefaultRate     float64         `json:"default_rate"`
	UnitLabel       string          `json:"unit_label"`
	TransactionType TransactionType `json:"transaction_type"`
}

type CalculateQuoteInput struct {
	BillingCategoryID string  `json:"billing_category_id"`
	Quantity          float64 `json:"quantity"`
	Rate              float64 `json:"rate"`
	Description       string  `json:"description"`
}

type QuoteResult struct {
	BillingCategoryID string          `json:"billing_category_id"`
	CategoryName      string          `json:"category_name"`
	CalcType          BillingCalcType `json:"calc_type"`
	UnitLabel         string          `json:"unit_label"`
	Quantity          float64         `json:"quantity"`
	Rate              float64         `json:"rate"`
	Amount            float64         `json:"amount"`
	Description       string          `json:"description"`
	SuggestedType     TransactionType `json:"suggested_type"`
}

type AccountingCategorySummary struct {
	CategoryID   string          `json:"category_id"`
	Name         string          `json:"name"`
	Type         TransactionType `json:"type"`
	Income       float64         `json:"income"`
	Expense      float64         `json:"expense"`
	Net          float64         `json:"net"`
	Transactions int             `json:"transactions"`
	LastActivity time.Time       `json:"last_activity"`
}

type DashboardSummary struct {
	Balance           Balance                     `json:"balance"`
	CategoryBreakdown []AccountingCategorySummary `json:"category_breakdown"`
}

type CategoryRankingMetric string

const (
	RankingByExpense      CategoryRankingMetric = "expense"
	RankingByIncome       CategoryRankingMetric = "income"
	RankingByTransactions CategoryRankingMetric = "transactions"
)

type CategoryRankingInput struct {
	Metric CategoryRankingMetric `json:"metric"`
	Days   int                   `json:"days"`
	Limit  int                   `json:"limit"`
}

type CategoryRankingItem struct {
	CategoryID   string  `json:"category_id"`
	Name         string  `json:"name"`
	Income       float64 `json:"income"`
	Expense      float64 `json:"expense"`
	Transactions int     `json:"transactions"`
}

type CategoryRankingResult struct {
	Items        []CategoryRankingItem `json:"items"`
	StartDate    string                `json:"start_date"`
	EndDate      string                `json:"end_date"`
	TotalIncome  float64               `json:"total_income"`
	TotalExpense float64               `json:"total_expense"`
}

type AppSettings struct {
	RankingDays  int     `json:"ranking_days"`
	WeekStartDay int     `json:"week_start_day"`
	WeeklyBudget float64 `json:"weekly_budget"`
}

type Budget struct {
	ID           string  `json:"id"`
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	WeeklyAmount float64 `json:"weekly_amount"`
	Active       bool    `json:"active"`
}

type SaveBudgetInput struct {
	CategoryID   string  `json:"category_id"`
	WeeklyAmount float64 `json:"weekly_amount"`
}

type WeeklyCategorySummary struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Spent        float64 `json:"spent"`
	Budget       float64 `json:"budget"`
	PercentUsed  float64 `json:"percent_used"`
	Remaining    float64 `json:"remaining"`
	DeltaVsPrev  float64 `json:"delta_vs_prev"`
}

type WeeklySummary struct {
	StartDate        string                  `json:"start_date"`
	EndDate          string                  `json:"end_date"`
	TotalSpent       float64                 `json:"total_spent"`
	TotalBudget      float64                 `json:"total_budget"`
	TotalPercent     float64                 `json:"total_percent"`
	TotalRemaining   float64                 `json:"total_remaining"`
	DeltaVsPrev      float64                 `json:"delta_vs_prev"`
	GlobalBudget     float64                 `json:"global_budget"`
	GlobalPercent    float64                 `json:"global_percent"`
	GlobalRemaining  float64                 `json:"global_remaining"`
	AvailableBalance float64                 `json:"available_balance"`
	Categories       []WeeklyCategorySummary `json:"categories"`
}

type AIProvider struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	BaseURL  string `json:"base_url"`
	ChatPath string `json:"chat_path"`
	APIKey   string `json:"api_key"`
	Active   bool   `json:"active"`
}

type AIModel struct {
	ID         string `json:"id"`
	ProviderID string `json:"provider_id"`
	Name       string `json:"name"`
	Active     bool   `json:"active"`
}

type AIConfiguration struct {
	ProviderID string `json:"provider_id"`
	ModelID    string `json:"model_id"`
}

type SaveAIConfigurationInput struct {
	ProviderID string `json:"provider_id"`
	ModelID    string `json:"model_id"`
	APIKey     string `json:"api_key"`
}

type CreateAIProviderInput struct {
	Name     string `json:"name"`
	BaseURL  string `json:"base_url"`
	ChatPath string `json:"chat_path"`
}

type CreateAIModelInput struct {
	ProviderID string `json:"provider_id"`
	Name       string `json:"name"`
}

type AnalyzeDashboardInput struct {
	Question string `json:"question"`
}

type AIAnalysisRecord struct {
	ID        string    `json:"id"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	Question  string    `json:"question"`
	Analysis  string    `json:"analysis"`
	CreatedAt time.Time `json:"created_at"`
}
