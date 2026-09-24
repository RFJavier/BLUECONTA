package backend

import (
	"context"
	"fmt"

	"contaduria/mvp/backend/bootstrap"
	"contaduria/mvp/backend/services"
	"contaduria/mvp/shared"
)

type App struct {
	ctx            context.Context
	runtime        *bootstrap.PocketBaseRuntime
	accounting     *services.AccountingService
	licenseManager *LicenseManager
}

func NewApp(runtime *bootstrap.PocketBaseRuntime) *App {
	return &App{
		runtime:        runtime,
		accounting:     services.NewAccountingService(runtime.App),
		licenseManager: NewLicenseManager(),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Shutdown(ctx context.Context) {
	_ = ctx
}

func (a *App) CreateCategory(input shared.CreateCategoryInput) (*shared.Category, error) {
	if !a.licenseManager.AllowsWriteOperations() {
		return nil, fmt.Errorf("license validation failed: write operations blocked")
	}
	return a.accounting.CreateCategory(input)
}

func (a *App) GetCategories() ([]shared.Category, error) {
	return a.accounting.GetCategories()
}

func (a *App) UpdateCategory(input shared.UpdateCategoryInput) (*shared.Category, error) {
	if !a.licenseManager.AllowsWriteOperations() {
		return nil, fmt.Errorf("license validation failed: write operations blocked")
	}
	return a.accounting.UpdateCategory(input)
}

func (a *App) DeleteCategory(categoryID string) error {
	if !a.licenseManager.AllowsWriteOperations() {
		return fmt.Errorf("license validation failed: write operations blocked")
	}
	return a.accounting.DeleteCategory(categoryID)
}

func (a *App) CreateTransaction(input shared.CreateTransactionInput) (*shared.Transaction, error) {
	if !a.licenseManager.AllowsWriteOperations() {
		return nil, fmt.Errorf("license validation failed: write operations blocked")
	}
	return a.accounting.CreateTransaction(input)
}

func (a *App) GetTransactions() ([]shared.Transaction, error) {
	return a.accounting.GetTransactions()
}

func (a *App) GetTransactionsFiltered(input shared.TransactionFilterInput) ([]shared.Transaction, error) {
	return a.accounting.GetTransactionsFiltered(input)
}

func (a *App) ExportTransactionsCSV(input shared.TransactionFilterInput) (string, error) {
	return a.accounting.ExportTransactionsCSV(input)
}

func (a *App) GetBalance() (*shared.Balance, error) {
	return a.accounting.GetBalance()
}

func (a *App) GetDashboardSummary() (*shared.DashboardSummary, error) {
	return a.accounting.GetDashboardSummary()
}

func (a *App) GetCategoryRanking(input shared.CategoryRankingInput) (*shared.CategoryRankingResult, error) {
	return a.accounting.GetCategoryRanking(input)
}

func (a *App) GetAppSettings() (*shared.AppSettings, error) {
	return a.accounting.GetAppSettings()
}

func (a *App) SaveAppSettings(input shared.AppSettings) (*shared.AppSettings, error) {
	if !a.licenseManager.AllowsWriteOperations() {
		return nil, fmt.Errorf("license validation failed: write operations blocked")
	}
	return a.accounting.SaveAppSettings(input)
}

func (a *App) GetAIProviders() ([]shared.AIProvider, error) {
	return a.accounting.GetAIProviders()
}

func (a *App) CreateAIProvider(input shared.CreateAIProviderInput) (*shared.AIProvider, error) {
	if !a.licenseManager.AllowsWriteOperations() {
		return nil, fmt.Errorf("license validation failed: write operations blocked")
	}
	return a.accounting.CreateAIProvider(input)
}

func (a *App) GetAIModels(providerID string) ([]shared.AIModel, error) {
	return a.accounting.GetAIModels(providerID)
}

func (a *App) CreateAIModel(input shared.CreateAIModelInput) (*shared.AIModel, error) {
	if !a.licenseManager.AllowsWriteOperations() {
		return nil, fmt.Errorf("license validation failed: write operations blocked")
	}
	return a.accounting.CreateAIModel(input)
}

func (a *App) GetAIConfiguration() (*shared.AIConfiguration, error) {
	return a.accounting.GetAIConfiguration()
}

func (a *App) SaveAIConfiguration(input shared.SaveAIConfigurationInput) (*shared.AIConfiguration, error) {
	if !a.licenseManager.AllowsWriteOperations() {
		return nil, fmt.Errorf("license validation failed: write operations blocked")
	}
	return a.accounting.SaveAIConfiguration(input)
}

func (a *App) AnalyzeDashboardWithAI(input shared.AnalyzeDashboardInput) (*shared.AIAnalysisRecord, error) {
	return a.accounting.AnalyzeDashboardWithAI(input)
}

func (a *App) GetAIAnalyses() ([]shared.AIAnalysisRecord, error) {
	return a.accounting.GetAIAnalyses()
}

func (a *App) DeleteAIAnalysis(id string) error {
	if !a.licenseManager.AllowsWriteOperations() {
		return fmt.Errorf("license validation failed: write operations blocked")
	}
	return a.accounting.DeleteAIAnalysis(id)
}

func (a *App) CreateBillingCategory(input shared.CreateBillingCategoryInput) (*shared.BillingCategory, error) {
	if !a.licenseManager.AllowsWriteOperations() {
		return nil, fmt.Errorf("license validation failed: write operations blocked")
	}
	return a.accounting.CreateBillingCategory(input)
}

func (a *App) GetBillingCategories() ([]shared.BillingCategory, error) {
	return a.accounting.GetBillingCategories()
}

func (a *App) UpdateBillingCategory(input shared.UpdateBillingCategoryInput) (*shared.BillingCategory, error) {
	if !a.licenseManager.AllowsWriteOperations() {
		return nil, fmt.Errorf("license validation failed: write operations blocked")
	}
	return a.accounting.UpdateBillingCategory(input)
}

func (a *App) DeleteBillingCategory(id string) error {
	if !a.licenseManager.AllowsWriteOperations() {
		return fmt.Errorf("license validation failed: write operations blocked")
	}
	return a.accounting.DeleteBillingCategory(id)
}

func (a *App) CalculateQuote(input shared.CalculateQuoteInput) (*shared.QuoteResult, error) {
	return a.accounting.CalculateQuote(input)
}

type AppStatus struct {
	LicenseStatus string `json:"license_status"`
}

func (a *App) GetAppStatus() AppStatus {
	return AppStatus{LicenseStatus: a.licenseManager.Status()}
}
