package bootstrap

import (
	"database/sql"
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type PocketBaseRuntime struct {
	App *pocketbase.PocketBase
}

func NewPocketBaseRuntime(dataDir string) (*PocketBaseRuntime, error) {
	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  dataDir,
		HideStartBanner: true,
	})

	if err := app.Bootstrap(); err != nil {
		return nil, fmt.Errorf("bootstrap pocketbase app: %w", err)
	}

	if err := ensureCollections(app); err != nil {
		return nil, fmt.Errorf("ensure collections: %w", err)
	}

	return &PocketBaseRuntime{App: app}, nil
}

func ensureCollections(app *pocketbase.PocketBase) error {
	categories, err := app.FindCollectionByNameOrId("categories")
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("find categories collection: %w", err)
		}

		categories = newCategoriesCollection()
		if err := app.Save(categories); err != nil {
			return fmt.Errorf("create categories collection: %w", err)
		}

		if err := seedCategories(app, categories); err != nil {
			return fmt.Errorf("seed categories: %w", err)
		}
	}

	if _, err := app.FindCollectionByNameOrId("transactions"); err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("find transactions collection: %w", err)
		}

		transactions := newTransactionsCollection(categories.Id)
		if err := app.Save(transactions); err != nil {
			return fmt.Errorf("create transactions collection: %w", err)
		}
	}

	if _, err := app.FindCollectionByNameOrId("budgets"); err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("find budgets collection: %w", err)
		}

		budgets := newBudgetsCollection(categories.Id)
		if err := app.Save(budgets); err != nil {
			return fmt.Errorf("create budgets collection: %w", err)
		}
	}

	if _, err := app.FindCollectionByNameOrId("billing_categories"); err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("find billing_categories collection: %w", err)
		}

		billingCategories := newBillingCategoriesCollection()
		if err := app.Save(billingCategories); err != nil {
			return fmt.Errorf("create billing_categories collection: %w", err)
		}

		if err := seedBillingCategories(app, billingCategories); err != nil {
			return fmt.Errorf("seed billing categories: %w", err)
		}
	}

	if _, err := app.FindCollectionByNameOrId("ai_providers"); err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("find ai_providers collection: %w", err)
		}

		aiProviders := newAIProvidersCollection()
		if err := app.Save(aiProviders); err != nil {
			return fmt.Errorf("create ai_providers collection: %w", err)
		}

		if err := seedAIProviders(app, aiProviders); err != nil {
			return fmt.Errorf("seed ai providers: %w", err)
		}
	}

	aiProviders, err := app.FindCollectionByNameOrId("ai_providers")
	if err != nil {
		return fmt.Errorf("load ai_providers collection: %w", err)
	}

	if _, err := app.FindCollectionByNameOrId("ai_models"); err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("find ai_models collection: %w", err)
		}

		aiModels := newAIModelsCollection(aiProviders.Id)
		if err := app.Save(aiModels); err != nil {
			return fmt.Errorf("create ai_models collection: %w", err)
		}

		if err := seedAIModels(app, aiProviders, aiModels); err != nil {
			return fmt.Errorf("seed ai models: %w", err)
		}
	}

	if _, err := app.FindCollectionByNameOrId("ai_settings"); err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("find ai_settings collection: %w", err)
		}

		aiSettings := newAISettingsCollection(aiProviders.Id)
		if err := app.Save(aiSettings); err != nil {
			return fmt.Errorf("create ai_settings collection: %w", err)
		}
	}

	if _, err := app.FindCollectionByNameOrId("ai_analyses"); err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("find ai_analyses collection: %w", err)
		}

		aiAnalyses := newAIAnalysesCollection()
		if err := app.Save(aiAnalyses); err != nil {
			return fmt.Errorf("create ai_analyses collection: %w", err)
		}
	}

	if _, err := app.FindCollectionByNameOrId("app_settings"); err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("find app_settings collection: %w", err)
		}

		appSettings := newAppSettingsCollection()
		if err := app.Save(appSettings); err != nil {
			return fmt.Errorf("create app_settings collection: %w", err)
		}
	}

	return nil
}

func newCategoriesCollection() *core.Collection {
	collection := core.NewBaseCollection("categories")
	collection.Fields.Add(
		&core.TextField{
			Name:     "name",
			Required: true,
			Max:      120,
		},
		&core.TextField{
			Name:     "type",
			Required: true,
			Pattern:  "^(income|expense)$",
			Max:      20,
		},
	)
	collection.AddIndex("idx_categories_name_type_unique", true, "name,type", "")

	return collection
}

func newTransactionsCollection(categoriesID string) *core.Collection {
	collection := core.NewBaseCollection("transactions")
	collection.Fields.Add(
		&core.TextField{
			Name:     "type",
			Required: true,
			Pattern:  "^(income|expense)$",
			Max:      20,
		},
		&core.NumberField{
			Name:     "amount",
			Required: true,
			Min:      float64Ptr(0),
		},
		&core.TextField{
			Name: "description",
			Max:  240,
		},
		&core.RelationField{
			Name:         "category_id",
			Required:     true,
			CollectionId: categoriesID,
			MaxSelect:    1,
		},
		&core.AutodateField{
			Name:     "created_at",
			OnCreate: true,
		},
	)
	collection.AddIndex("idx_transactions_created_at", false, "created_at", "")

	return collection
}

func newBudgetsCollection(categoriesID string) *core.Collection {
	collection := core.NewBaseCollection("budgets")
	collection.Fields.Add(
		&core.RelationField{
			Name:         "category_id",
			Required:     true,
			CollectionId: categoriesID,
			MaxSelect:    1,
		},
		&core.NumberField{
			Name:     "weekly_amount",
			Required: true,
			Min:      float64Ptr(0),
		},
		&core.BoolField{
			Name: "active",
		},
	)
	collection.AddIndex("idx_budgets_category_unique", true, "category_id", "")

	return collection
}

func newBillingCategoriesCollection() *core.Collection {
	collection := core.NewBaseCollection("billing_categories")
	collection.Fields.Add(
		&core.TextField{
			Name:     "name",
			Required: true,
			Max:      120,
		},
		&core.TextField{
			Name:     "calc_type",
			Required: true,
			Pattern:  "^(hourly|unit)$",
			Max:      20,
		},
		&core.NumberField{
			Name:     "default_rate",
			Required: true,
			Min:      float64Ptr(0),
		},
		&core.TextField{
			Name:     "unit_label",
			Required: true,
			Max:      60,
		},
		&core.TextField{
			Name:     "transaction_type",
			Required: true,
			Pattern:  "^(income|expense)$",
			Max:      20,
		},
	)
	collection.AddIndex("idx_billing_categories_unique", true, "name,calc_type", "")
	return collection
}

func seedBillingCategories(app *pocketbase.PocketBase, collection *core.Collection) error {
	seed := []struct {
		name            string
		calcType        string
		defaultRate     float64
		unitLabel       string
		transactionType string
	}{
		{name: "Servicios profesionales", calcType: "hourly", defaultRate: 25, unitLabel: "hora", transactionType: "income"},
		{name: "Asesoria", calcType: "hourly", defaultRate: 18, unitLabel: "hora", transactionType: "income"},
		{name: "Digitalizacion de documentos", calcType: "unit", defaultRate: 0.80, unitLabel: "folio", transactionType: "income"},
	}

	for _, item := range seed {
		record := core.NewRecord(collection)
		record.Set("name", item.name)
		record.Set("calc_type", item.calcType)
		record.Set("default_rate", item.defaultRate)
		record.Set("unit_label", item.unitLabel)
		record.Set("transaction_type", item.transactionType)
		if err := app.Save(record); err != nil {
			return err
		}
	}

	return nil
}

func seedCategories(app *pocketbase.PocketBase, collection *core.Collection) error {
	seeds := []struct {
		name string
		kind string
	}{
		{name: "Fondos personales", kind: "income"},
		{name: "Alimentación", kind: "expense"},
		{name: "Transporte", kind: "expense"},
		{name: "Otros", kind: "expense"},
	}

	for _, seed := range seeds {
		record := core.NewRecord(collection)
		record.Set("name", seed.name)
		record.Set("type", seed.kind)
		if err := app.Save(record); err != nil {
			return err
		}
	}

	return nil
}

func newAIProvidersCollection() *core.Collection {
	collection := core.NewBaseCollection("ai_providers")
	collection.Fields.Add(
		&core.TextField{
			Name:     "name",
			Required: true,
			Max:      100,
		},
		&core.TextField{
			Name:     "base_url",
			Required: true,
			Max:      240,
		},
		&core.TextField{
			Name:     "chat_path",
			Required: true,
			Max:      120,
		},
		&core.TextField{
			Name: "api_key",
			Max:  300,
		},
		&core.BoolField{
			Name: "active",
		},
	)
	collection.AddIndex("idx_ai_providers_name_unique", true, "name", "")
	return collection
}

func newAIModelsCollection(providerCollectionID string) *core.Collection {
	collection := core.NewBaseCollection("ai_models")
	collection.Fields.Add(
		&core.RelationField{
			Name:         "provider_id",
			Required:     true,
			CollectionId: providerCollectionID,
			MaxSelect:    1,
		},
		&core.TextField{
			Name:     "name",
			Required: true,
			Max:      120,
		},
		&core.BoolField{
			Name: "active",
		},
	)
	collection.AddIndex("idx_ai_models_provider_name_unique", true, "provider_id,name", "")
	return collection
}

func newAISettingsCollection(providerCollectionID string) *core.Collection {
	collection := core.NewBaseCollection("ai_settings")
	collection.Fields.Add(
		&core.RelationField{
			Name:         "provider_id",
			CollectionId: providerCollectionID,
			MaxSelect:    1,
		},
		&core.TextField{
			Name: "model_id",
			Max:  40,
		},
	)
	return collection
}

func newAIAnalysesCollection() *core.Collection {
	collection := core.NewBaseCollection("ai_analyses")
	collection.Fields.Add(
		&core.TextField{
			Name: "question",
			Max:  280,
		},
		&core.TextField{
			Name:     "provider",
			Required: true,
			Max:      100,
		},
		&core.TextField{
			Name:     "model",
			Required: true,
			Max:      120,
		},
		&core.TextField{
			Name: "analysis",
			Max:  0,
		},
		&core.AutodateField{
			Name:     "created_at",
			OnCreate: true,
		},
	)
	collection.AddIndex("idx_ai_analyses_created_at", false, "created_at", "")
	return collection
}

func newAppSettingsCollection() *core.Collection {
	collection := core.NewBaseCollection("app_settings")
	collection.Fields.Add(
		&core.NumberField{
			Name: "ranking_days",
			Min:  float64Ptr(0),
		},
		&core.NumberField{
			Name:    "week_start_day",
			Min:     float64Ptr(0),
			Max:     float64Ptr(7),
			OnlyInt: true,
		},
		&core.NumberField{
			Name: "weekly_budget",
			Min:  float64Ptr(0),
		},
		&core.BoolField{
			Name: "allow_over_budget",
		},
		&core.BoolField{
			Name: "allow_over_global",
		},
	)
	return collection
}

func seedAIProviders(app *pocketbase.PocketBase, collection *core.Collection) error {
	record := core.NewRecord(collection)
	record.Set("name", "DeepSeek")
	record.Set("base_url", "https://api.deepseek.com")
	record.Set("chat_path", "/chat/completions")
	record.Set("active", true)
	return app.Save(record)
}

func seedAIModels(app *pocketbase.PocketBase, providersCollection *core.Collection, modelsCollection *core.Collection) error {
	providers, err := app.FindAllRecords(providersCollection.Name)
	if err != nil || len(providers) == 0 {
		return err
	}

	record := core.NewRecord(modelsCollection)
	record.Set("provider_id", providers[0].Id)
	record.Set("name", "deepseek-chat")
	record.Set("active", true)
	return app.Save(record)
}

func float64Ptr(v float64) *float64 {
	return &v
}
