export namespace backend {
	
	export class AppStatus {
	    license_status: string;
	
	    static createFrom(source: any = {}) {
	        return new AppStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.license_status = source["license_status"];
	    }
	}

}

export namespace shared {
	
	export class AIAnalysisRecord {
	    id: string;
	    provider: string;
	    model: string;
	    question: string;
	    analysis: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new AIAnalysisRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.question = source["question"];
	        this.analysis = source["analysis"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AIConfiguration {
	    provider_id: string;
	    model_id: string;
	
	    static createFrom(source: any = {}) {
	        return new AIConfiguration(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider_id = source["provider_id"];
	        this.model_id = source["model_id"];
	    }
	}
	export class AIModel {
	    id: string;
	    provider_id: string;
	    name: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AIModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.provider_id = source["provider_id"];
	        this.name = source["name"];
	        this.active = source["active"];
	    }
	}
	export class AIProvider {
	    id: string;
	    name: string;
	    base_url: string;
	    chat_path: string;
	    api_key: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AIProvider(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.base_url = source["base_url"];
	        this.chat_path = source["chat_path"];
	        this.api_key = source["api_key"];
	        this.active = source["active"];
	    }
	}
	export class AccountingCategorySummary {
	    category_id: string;
	    name: string;
	    type: string;
	    income: number;
	    expense: number;
	    net: number;
	    transactions: number;
	    // Go type: time
	    last_activity: any;
	
	    static createFrom(source: any = {}) {
	        return new AccountingCategorySummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category_id = source["category_id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.income = source["income"];
	        this.expense = source["expense"];
	        this.net = source["net"];
	        this.transactions = source["transactions"];
	        this.last_activity = this.convertValues(source["last_activity"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AnalyzeDashboardInput {
	    question: string;
	
	    static createFrom(source: any = {}) {
	        return new AnalyzeDashboardInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.question = source["question"];
	    }
	}
	export class AppSettings {
	    ranking_days: number;
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ranking_days = source["ranking_days"];
	    }
	}
	export class Balance {
	    income: number;
	    expense: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new Balance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.income = source["income"];
	        this.expense = source["expense"];
	        this.total = source["total"];
	    }
	}
	export class BillingCategory {
	    id: string;
	    name: string;
	    calc_type: string;
	    default_rate: number;
	    unit_label: string;
	    transaction_type: string;
	
	    static createFrom(source: any = {}) {
	        return new BillingCategory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.calc_type = source["calc_type"];
	        this.default_rate = source["default_rate"];
	        this.unit_label = source["unit_label"];
	        this.transaction_type = source["transaction_type"];
	    }
	}
	export class CalculateQuoteInput {
	    billing_category_id: string;
	    quantity: number;
	    rate: number;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new CalculateQuoteInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.billing_category_id = source["billing_category_id"];
	        this.quantity = source["quantity"];
	        this.rate = source["rate"];
	        this.description = source["description"];
	    }
	}
	export class Category {
	    id: string;
	    name: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new Category(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	    }
	}
	export class CategoryRankingInput {
	    metric: string;
	    days: number;
	    limit: number;
	
	    static createFrom(source: any = {}) {
	        return new CategoryRankingInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.metric = source["metric"];
	        this.days = source["days"];
	        this.limit = source["limit"];
	    }
	}
	export class CategoryRankingItem {
	    category_id: string;
	    name: string;
	    income: number;
	    expense: number;
	    transactions: number;
	
	    static createFrom(source: any = {}) {
	        return new CategoryRankingItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category_id = source["category_id"];
	        this.name = source["name"];
	        this.income = source["income"];
	        this.expense = source["expense"];
	        this.transactions = source["transactions"];
	    }
	}
	export class CategoryRankingResult {
	    items: CategoryRankingItem[];
	    start_date: string;
	    end_date: string;
	    total_income: number;
	    total_expense: number;
	
	    static createFrom(source: any = {}) {
	        return new CategoryRankingResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], CategoryRankingItem);
	        this.start_date = source["start_date"];
	        this.end_date = source["end_date"];
	        this.total_income = source["total_income"];
	        this.total_expense = source["total_expense"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CreateAIModelInput {
	    provider_id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateAIModelInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider_id = source["provider_id"];
	        this.name = source["name"];
	    }
	}
	export class CreateAIProviderInput {
	    name: string;
	    base_url: string;
	    chat_path: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateAIProviderInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.base_url = source["base_url"];
	        this.chat_path = source["chat_path"];
	    }
	}
	export class CreateBillingCategoryInput {
	    name: string;
	    calc_type: string;
	    default_rate: number;
	    unit_label: string;
	    transaction_type: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateBillingCategoryInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.calc_type = source["calc_type"];
	        this.default_rate = source["default_rate"];
	        this.unit_label = source["unit_label"];
	        this.transaction_type = source["transaction_type"];
	    }
	}
	export class CreateCategoryInput {
	    name: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateCategoryInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	    }
	}
	export class CreateTransactionInput {
	    type: string;
	    amount: number;
	    description: string;
	    category_id: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateTransactionInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.amount = source["amount"];
	        this.description = source["description"];
	        this.category_id = source["category_id"];
	    }
	}
	export class DashboardSummary {
	    balance: Balance;
	    category_breakdown: AccountingCategorySummary[];
	
	    static createFrom(source: any = {}) {
	        return new DashboardSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.balance = this.convertValues(source["balance"], Balance);
	        this.category_breakdown = this.convertValues(source["category_breakdown"], AccountingCategorySummary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class QuoteResult {
	    billing_category_id: string;
	    category_name: string;
	    calc_type: string;
	    unit_label: string;
	    quantity: number;
	    rate: number;
	    amount: number;
	    description: string;
	    suggested_type: string;
	
	    static createFrom(source: any = {}) {
	        return new QuoteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.billing_category_id = source["billing_category_id"];
	        this.category_name = source["category_name"];
	        this.calc_type = source["calc_type"];
	        this.unit_label = source["unit_label"];
	        this.quantity = source["quantity"];
	        this.rate = source["rate"];
	        this.amount = source["amount"];
	        this.description = source["description"];
	        this.suggested_type = source["suggested_type"];
	    }
	}
	export class SaveAIConfigurationInput {
	    provider_id: string;
	    model_id: string;
	    api_key: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveAIConfigurationInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider_id = source["provider_id"];
	        this.model_id = source["model_id"];
	        this.api_key = source["api_key"];
	    }
	}
	export class Transaction {
	    id: string;
	    type: string;
	    amount: number;
	    description: string;
	    category_id: string;
	    category: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Transaction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.amount = source["amount"];
	        this.description = source["description"];
	        this.category_id = source["category_id"];
	        this.category = source["category"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TransactionFilterInput {
	    start_date: string;
	    end_date: string;
	    category_id: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new TransactionFilterInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.start_date = source["start_date"];
	        this.end_date = source["end_date"];
	        this.category_id = source["category_id"];
	        this.type = source["type"];
	    }
	}
	export class UpdateBillingCategoryInput {
	    id: string;
	    name: string;
	    calc_type: string;
	    default_rate: number;
	    unit_label: string;
	    transaction_type: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateBillingCategoryInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.calc_type = source["calc_type"];
	        this.default_rate = source["default_rate"];
	        this.unit_label = source["unit_label"];
	        this.transaction_type = source["transaction_type"];
	    }
	}
	export class UpdateCategoryInput {
	    id: string;
	    name: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCategoryInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	    }
	}

}

