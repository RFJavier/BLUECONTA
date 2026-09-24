const state = {
  activeView: "dashboard",
  categories: [],
  transactions: [],
  billingCategories: [],
  aiProviders: [],
  aiModels: [],
  aiConfig: null,
  aiAnalysis: null,
  aiAnalyses: [],
  quoteResult: null,
  dashboardSummary: null,
  rankingIncome: [],
  rankingExpense: [],
  rankingTotalIncome: 0,
  rankingTotalExpense: 0,
  rankingStartDate: "",
  rankingEndDate: "",
  rankingMetric: "income",
  appSettings: { ranking_days: 0 },
  transactionFilters: {
    start_date: "",
    end_date: "",
    category_id: "",
    type: "",
  },
};

const toastEl = document.getElementById("toast");
const quoteBillingCategoryEl = document.getElementById("quoteBillingCategory");
const quoteRateEl = document.getElementById("quoteRate");
const quoteUnitLabelEl = document.getElementById("quoteUnitLabel");
const filterStartDateEl = document.getElementById("filterStartDate");
const filterEndDateEl = document.getElementById("filterEndDate");
const filterCategoryEl = document.getElementById("filterCategory");
const filterTypeEl = document.getElementById("filterType");
const aiProviderSelectEl = document.getElementById("aiProviderSelect");
const aiModelSelectEl = document.getElementById("aiModelSelect");
const aiApiKeyEl = document.getElementById("aiApiKey");
const newModelProviderEl = document.getElementById("newModelProvider");

const categoryModalEl = document.getElementById("categoryModal");
const transactionModalEl = document.getElementById("transactionModal");
const transactionModalCategorySelectEl = document.getElementById("transactionModalCategorySelect");

let categoryModalType = "income";
const txModalState = { categoryId: null, type: "income" };

function fmtMoney(value) {
  return new Intl.NumberFormat("es-SV", { style: "currency", currency: "USD" }).format(value || 0);
}

function fmtDate(raw) {
  if (!raw) return "-";
  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat("es-SV", { dateStyle: "short", timeStyle: "short" }).format(date);
}

function fmtDateShort(raw) {
  if (!raw) return "-";
  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat("es-SV", { dateStyle: "short" }).format(date);
}

function fmtRangeDate(iso) {
  if (!iso) return "-";
  const parts = String(iso).split("-");
  if (parts.length !== 3) return iso;
  return `${parts[2]}/${parts[1]}/${parts[0]}`;
}

function esc(value) {
  return String(value || "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function showToast(message, isError = false) {
  toastEl.textContent = message;
  toastEl.classList.add("visible");
  toastEl.classList.toggle("error", isError);
  clearTimeout(showToast._timer);
  showToast._timer = setTimeout(() => toastEl.classList.remove("visible"), 2800);
}

function getBackend() {
  if (!window.go || !window.go.backend || !window.go.backend.App) {
    throw new Error("Wails backend no esta disponible");
  }
  return window.go.backend.App;
}

function getCategoryById(id) {
  return state.categories.find((item) => item.id === id);
}

function getBillingCategoryById(id) {
  return state.billingCategories.find((item) => item.id === id);
}

function getAIProviderById(id) {
  return state.aiProviders.find((item) => item.id === id);
}

function readTransactionFiltersFromUI() {
  return {
    start_date: filterStartDateEl.value || "",
    end_date: filterEndDateEl.value || "",
    category_id: filterCategoryEl.value || "",
    type: filterTypeEl.value || "",
  };
}

function switchView(viewName) {
  state.activeView = viewName;

  document.querySelectorAll(".view").forEach((view) => {
    const isTarget = view.id === `view-${viewName}`;
    view.classList.toggle("hidden-view", !isTarget);
  });

  document.querySelectorAll(".menu-tile").forEach((button) => {
    const isTarget = button.dataset.view === viewName;
    button.classList.toggle("active", isTarget);
  });

  closeMenu();
}

function openMenu() {
  document.getElementById("menuDropdown").classList.remove("hidden");
  document.getElementById("menuBtn").setAttribute("aria-expanded", "true");
}

function closeMenu() {
  document.getElementById("menuDropdown").classList.add("hidden");
  document.getElementById("menuBtn").setAttribute("aria-expanded", "false");
}

function toggleMenu() {
  const dropdown = document.getElementById("menuDropdown");
  if (dropdown.classList.contains("hidden")) {
    openMenu();
  } else {
    closeMenu();
  }
}

function renderFilterCategoryOptions() {
  const current = filterCategoryEl.value;
  filterCategoryEl.innerHTML = `<option value="">Todas</option>${state.categories
    .map((c) => `<option value='${c.id}'>${esc(c.name)}</option>`)
    .join("")}`;

  if (current && state.categories.some((c) => c.id === current)) {
    filterCategoryEl.value = current;
  } else {
    filterCategoryEl.value = state.transactionFilters.category_id || "";
  }
}

function renderBillingCategories() {
  const list = document.getElementById("billingCategoriesList");
  if (state.billingCategories.length === 0) {
    list.innerHTML = "<li class='empty'>No hay categorias de cobro aun.</li>";
    quoteBillingCategoryEl.innerHTML = "<option value=''>Crea una categoria de cobro primero</option>";
    return;
  }

  list.innerHTML = state.billingCategories
    .map((c) => {
      const calcTypeLabel = c.calc_type === "hourly" ? "Por hora" : "Por unidad";
      const txTypeLabel = c.transaction_type === "income" ? "Ingreso" : "Gasto";
      return `<li>
        <div>
          <strong>${esc(c.name)}</strong>
          <small>${calcTypeLabel} | ${esc(c.unit_label)} | ${txTypeLabel} | ${fmtMoney(c.default_rate)}</small>
        </div>
        <div class="actions">
          <button type="button" class="tiny secondary" data-action="edit-billing" data-id="${c.id}">Editar</button>
          <button type="button" class="tiny danger" data-action="delete-billing" data-id="${c.id}">Eliminar</button>
        </div>
      </li>`;
    })
    .join("");

  quoteBillingCategoryEl.innerHTML = state.billingCategories
    .map((c) => `<option value='${c.id}'>${esc(c.name)}</option>`)
    .join("");

  syncQuoteFieldsFromBillingCategory();
}

function renderTransactions() {
  const tbody = document.getElementById("transactionsBody");
  if (state.transactions.length === 0) {
    tbody.innerHTML = "<tr><td colspan='5' class='empty'>Sin transacciones para el filtro seleccionado.</td></tr>";
    return;
  }

  tbody.innerHTML = state.transactions
    .map((tx) => {
      const sign = tx.type === "expense" ? "-" : "+";
      return `<tr>
        <td>${fmtDate(tx.created_at)}</td>
        <td>${tx.type === "income" ? "Ingreso" : "Gasto"}</td>
        <td>${esc(tx.category || "-")}</td>
        <td>${esc(tx.description || "-")}</td>
        <td class='amount ${tx.type}'>${sign}${fmtMoney(tx.amount)}</td>
      </tr>`;
    })
    .join("");
}

function renderDashboard() {
  const summary = state.dashboardSummary;
  if (!summary) return;

  renderBalanceDonut(summary.balance);
  renderCategoryCards();

  const tbody = document.getElementById("dashboardBillingBody");
  const rows = summary.category_breakdown || [];

  if (rows.length === 0) {
    tbody.innerHTML = "<tr><td colspan='4' class='empty'>No hay movimientos en categorias contables.</td></tr>";
    return;
  }

  tbody.innerHTML = rows
    .map(
      (row) =>
        `<tr>
          <td>${esc(row.name)}</td>
          <td>${fmtMoney(row.income)}</td>
          <td>${fmtMoney(row.expense)}</td>
          <td>${fmtMoney(row.net)}</td>
        </tr>`
    )
    .join("");
}

function renderAIConfiguration() {
  const providerOptions =
    state.aiProviders.length === 0
      ? "<option value=''>No hay proveedores</option>"
      : state.aiProviders.map((p) => `<option value='${p.id}'>${esc(p.name)}</option>`).join("");

  aiProviderSelectEl.innerHTML = providerOptions;
  newModelProviderEl.innerHTML = providerOptions;

  const selectedProvider = state.aiConfig?.provider_id || state.aiProviders[0]?.id || "";
  aiProviderSelectEl.value = selectedProvider;
  newModelProviderEl.value = selectedProvider;
  renderAIModelOptions(selectedProvider);
}

function renderAIModelOptions(providerID) {
  const models = state.aiModels.filter((m) => m.provider_id === providerID);
  if (models.length === 0) {
    aiModelSelectEl.innerHTML = "<option value=''>No hay modelos para este proveedor</option>";
    return;
  }

  aiModelSelectEl.innerHTML = models.map((m) => `<option value='${m.id}'>${esc(m.name)}</option>`).join("");
  aiModelSelectEl.value = state.aiConfig?.model_id && models.some((m) => m.id === state.aiConfig.model_id) ? state.aiConfig.model_id : models[0].id;
}

function renderAIAnalysis() {
  const box = document.getElementById("aiAnalysisBox");
  if (!state.aiAnalysis) {
    box.classList.add("hidden");
    return;
  }

  box.classList.remove("hidden");
  document.getElementById("aiAnalysisMeta").textContent = `${state.aiAnalysis.provider} | ${state.aiAnalysis.model} | ${fmtDate(state.aiAnalysis.created_at)}`;
  document.getElementById("aiAnalysisQuestion").textContent = state.aiAnalysis.question ? `Pregunta: ${state.aiAnalysis.question}` : "";
  document.getElementById("aiAnalysisText").textContent = state.aiAnalysis.analysis;
}

function renderAIAnalyses() {
  const list = document.getElementById("aiAnalysesList");
  if (!state.aiAnalyses.length) {
    list.innerHTML = "<p class='empty'>No hay analisis guardados aun.</p>";
    return;
  }

  list.innerHTML = state.aiAnalyses
    .map((a) => {
      const question = a.question ? esc(a.question) : "Sin pregunta";
      return `<div class="ai-analysis-card" data-id="${a.id}">
        <div class="ai-analysis-card-head">
          <div>
            <strong>${esc(a.model)}</strong>
            <small>${fmtDate(a.created_at)} | ${esc(a.provider)}</small>
          </div>
          <div class="actions">
            <button type="button" class="tiny secondary" data-action="view-analysis" data-id="${a.id}">Ver</button>
            <button type="button" class="tiny danger" data-action="delete-analysis" data-id="${a.id}">Eliminar</button>
          </div>
        </div>
        <p class="ai-analysis-question muted">${question}</p>
      </div>`;
    })
    .join("");
}

function renderBalanceDonut(balance) {
  const income = Number(balance.income || 0);
  const expense = Number(balance.expense || 0);
  const flow = income + expense;
  const net = income - expense;
  const incomePct = flow > 0 ? (income / flow) * 100 : 50;

  const donutEl = document.getElementById("balanceDonut");
  donutEl.style.setProperty("--income-pct", `${incomePct}%`);

  const balanceEl = document.getElementById("donutBalance");
  balanceEl.textContent = fmtMoney(net);
  balanceEl.classList.toggle("income", net >= 0);
  balanceEl.classList.toggle("expense", net < 0);

  document.getElementById("legendIncome").textContent = fmtMoney(income);
  document.getElementById("legendExpense").textContent = fmtMoney(expense);

  const flowEl = document.getElementById("donutFlow");
  flowEl.textContent = fmtMoney(flow);

  const flowIndicator = document.getElementById("flowIndicator");
  flowIndicator.classList.toggle("income", net >= 0);
  flowIndicator.classList.toggle("expense", net < 0);
}

function renderCategoryCards() {
  const container = document.getElementById("categoryBars");
  const categories = state.categories || [];

  if (!categories.length) {
    container.innerHTML = "<p class='empty'>No hay categorias aun. Usa el boton 'Crear categoria'.</p>";
    return;
  }

  const breakdown = state.dashboardSummary?.category_breakdown || [];
  const byId = new Map(breakdown.map((b) => [b.category_id, b]));

  const amountOf = (c) => {
    const s = byId.get(c.id);
    return c.type === "income" ? Number(s?.income || 0) : Number(s?.expense || 0);
  };

  const buildCards = (list) => {
    const colorClass = list[0].type === "income" ? "income" : "expense";
    const typeLabel = colorClass === "income" ? "Ingreso" : "Egreso";
    const total = list.reduce((sum, c) => sum + amountOf(c), 0);

    const sorted = [...list].sort((a, b) => amountOf(b) - amountOf(a));

    return sorted
      .map((c) => {
        const s = byId.get(c.id);
        const amount = amountOf(c);
        const count = s?.transactions || 0;
        const last = s?.last_activity ? fmtDateShort(s.last_activity) : "";
        const pct = total > 0 ? (amount / total) * 100 : 0;
        const hasActivity = count > 0;

        const meta = hasActivity
          ? `${count} transacciones · ${last}`
          : "Sin movimientos";

        return `
          <div class="cat-card ${colorClass}${hasActivity ? "" : " is-empty"}" data-id="${c.id}" data-name="${esc(c.name)}">
            <div class="cat-top">
              <div class="cat-heading">
                <strong>${esc(c.name)}</strong>
                <span class="cat-pill ${colorClass}">${typeLabel}</span>
              </div>
              <span class="cat-value ${colorClass}">${fmtMoney(amount)}</span>
            </div>
            <div class="cat-track">
              <span class="cat-track-fill ${colorClass}" style="width:${pct}%"></span>
            </div>
            <div class="cat-bottom">
              <span class="cat-sub">${meta}</span>
              <span class="cat-add-btn ${colorClass}">+ Agregar</span>
            </div>
          </div>
        `;
      })
      .join("");
  };

  const incomeCats = categories.filter((c) => c.type === "income");
  const expenseCats = categories.filter((c) => c.type === "expense");

  let html = "";
  if (incomeCats.length) {
    html += `<div class="cat-group">
      <h4 class="cat-group-title income">Ingresos</h4>
      <div class="cat-grid">${buildCards(incomeCats)}</div>
    </div>`;
  }
  if (expenseCats.length) {
    html += `<div class="cat-group">
      <h4 class="cat-group-title expense">Egresos</h4>
      <div class="cat-grid">${buildCards(expenseCats)}</div>
    </div>`;
  }

  container.innerHTML = html;
}

function renderQuoteResult() {
  const wrapper = document.getElementById("quoteResult");
  if (!state.quoteResult) {
    wrapper.classList.add("hidden");
    return;
  }

  wrapper.classList.remove("hidden");
  document.getElementById("quoteAmount").textContent = fmtMoney(state.quoteResult.amount);
  document.getElementById("quoteSummary").textContent = `${state.quoteResult.quantity} ${state.quoteResult.unit_label} x ${fmtMoney(state.quoteResult.rate)} = ${fmtMoney(state.quoteResult.amount)}`;
}

function vbarColumns(items, colorClass) {
  if (!items.length) {
    return "<p class='empty'>Sin movimientos en el periodo.</p>";
  }

  const field = colorClass === "income" ? "income" : "expense";
  const label = colorClass === "income" ? "Ingresos" : "Egresos";
  const values = items.map((item) => Number(item[field] || 0));
  const max = Math.max(...values, 1);

  return items
    .map((item) => {
      const value = Number(item[field] || 0);
      const height = (value / max) * 100;

      return `<div class="vbar-col">
        <div class="vbar-pair">
          <span class="vbar ${colorClass}" style="height:${height}%" title="${label} ${fmtMoney(value)}"></span>
        </div>
        <span class="vbar-name" title="${esc(item.name)}">${esc(item.name)}</span>
      </div>`;
    })
    .join("");
}

function renderCategoryRanking() {
  const container = document.getElementById("categoryRanking");
  const metric = state.rankingMetric;

  if (metric === "mixed") {
    const incomeItems = state.rankingIncome;
    const expenseItems = state.rankingExpense;

    if (!incomeItems.length && !expenseItems.length) {
      container.innerHTML = "<p class='empty'>Sin movimientos en el periodo.</p>";
    } else {
      container.innerHTML = `
        <div class="vbar-mixed">
          <div class="vbar-group">
            <div class="vbar-group-title income">Ingresos</div>
            <div class="vbar-chart">${incomeItems.length ? vbarColumns(incomeItems, "income") : "<p class='empty'>Sin ingresos</p>"}</div>
          </div>
          <div class="vbar-group">
            <div class="vbar-group-title expense">Egresos</div>
            <div class="vbar-chart">${expenseItems.length ? vbarColumns(expenseItems, "expense") : "<p class='empty'>Sin egresos</p>"}</div>
          </div>
        </div>
      `;
    }
  } else {
    const colorClass = metric === "expense" ? "expense" : "income";
    const items = metric === "expense" ? state.rankingExpense : state.rankingIncome;
    container.innerHTML = `<div class="vbar-chart">${vbarColumns(items, colorClass)}</div>`;
  }

  renderRankingIndicator();
}

function renderRankingIndicator() {
  const el = document.getElementById("rankingIndicator");
  const metric = state.rankingMetric;
  const range = `${fmtRangeDate(state.rankingStartDate)} - ${fmtRangeDate(state.rankingEndDate)}`;

  const incomeItem = `<span class="indicator-item income">
    <span class="legend-dot income"></span>
    <span>Ingresos <strong>${fmtMoney(state.rankingTotalIncome)}</strong></span>
  </span>`;
  const expenseItem = `<span class="indicator-item expense">
    <span class="legend-dot expense"></span>
    <span>Egresos <strong>${fmtMoney(state.rankingTotalExpense)}</strong></span>
  </span>`;
  const rangeItem = `<span class="indicator-range">${range}</span>`;

  let content;
  if (metric === "income") {
    content = incomeItem;
  } else if (metric === "expense") {
    content = expenseItem;
  } else {
    content = incomeItem + expenseItem;
  }

  el.innerHTML = content + rangeItem;
}

function syncQuoteFieldsFromBillingCategory() {
  const selected = getBillingCategoryById(quoteBillingCategoryEl.value);
  if (!selected) return;

  quoteUnitLabelEl.textContent = selected.unit_label || (selected.calc_type === "hourly" ? "hora" : "unidad");
  if (!quoteRateEl.value || Number(quoteRateEl.value) <= 0) {
    quoteRateEl.value = Number(selected.default_rate || 0).toFixed(2);
  }
}

async function refreshCategories() {
  state.categories = await getBackend().GetCategories();
  renderCategoryCards();
  renderFilterCategoryOptions();
}

async function refreshTransactions() {
  state.transactions = await getBackend().GetTransactionsFiltered(state.transactionFilters);
  renderTransactions();
}

async function refreshBillingCategories() {
  state.billingCategories = await getBackend().GetBillingCategories();
  renderBillingCategories();
}

async function refreshDashboard() {
  state.dashboardSummary = await getBackend().GetDashboardSummary();
  renderDashboard();
}

async function refreshCategoryRanking() {
  const [income, expense] = await Promise.all([
    getBackend().GetCategoryRanking({ metric: "income", limit: 5 }),
    getBackend().GetCategoryRanking({ metric: "expense", limit: 5 }),
  ]);
  state.rankingIncome = income.items;
  state.rankingExpense = expense.items;
  state.rankingTotalIncome = income.total_income;
  state.rankingTotalExpense = expense.total_expense;
  state.rankingStartDate = income.start_date;
  state.rankingEndDate = income.end_date;
  renderCategoryRanking();
}

async function refreshAIProviders() {
  state.aiProviders = await getBackend().GetAIProviders();
}

async function refreshAIModels() {
  state.aiModels = await getBackend().GetAIModels("");
}

async function refreshAIConfiguration() {
  state.aiConfig = await getBackend().GetAIConfiguration();
  renderAIConfiguration();
}

async function refreshAIAnalyses() {
  state.aiAnalyses = await getBackend().GetAIAnalyses();
  renderAIAnalyses();
}

async function refreshAppSettings() {
  state.appSettings = await getBackend().GetAppSettings();
  document.getElementById("appSettingsDays").value = state.appSettings.ranking_days ?? 0;
}

async function refreshAll() {
  await Promise.all([
    refreshCategories(),
    refreshTransactions(),
    refreshBillingCategories(),
    refreshDashboard(),
    refreshCategoryRanking(),
    refreshAIProviders(),
    refreshAIModels(),
    refreshAIAnalyses(),
    refreshAppSettings(),
  ]);
  await refreshAIConfiguration();
}

function openCategoryModal() {
  categoryModalType = "income";
  syncCategoryTypeButtons();
  document.getElementById("categoryModalName").value = "";
  categoryModalEl.classList.remove("hidden");
  document.getElementById("categoryModalName").focus();
}

function closeCategoryModal() {
  categoryModalEl.classList.add("hidden");
}

function syncCategoryTypeButtons() {
  document.querySelectorAll(".type-btn").forEach((btn) => {
    btn.classList.toggle("active", btn.dataset.type === categoryModalType);
  });
}

function onCategoryTypeClick(event) {
  const btn = event.target.closest(".type-btn");
  if (!btn) return;
  categoryModalType = btn.dataset.type;
  syncCategoryTypeButtons();
}

async function onCreateCategoryFromModal(event) {
  event.preventDefault();
  const name = document.getElementById("categoryModalName").value.trim();

  try {
    await getBackend().CreateCategory({ name, type: categoryModalType });
    closeCategoryModal();
    await Promise.all([refreshCategories(), refreshTransactions(), refreshDashboard(), refreshCategoryRanking()]);
    showToast("Categoria creada.");
  } catch (error) {
    showToast(error.message || "No se pudo crear la categoria.", true);
  }
}

async function onCreateBillingCategory(event) {
  event.preventDefault();

  const payload = {
    name: document.getElementById("billingName").value.trim(),
    calc_type: document.getElementById("billingCalcType").value,
    default_rate: Number(document.getElementById("billingRate").value),
    unit_label: document.getElementById("billingUnitLabel").value.trim(),
    transaction_type: document.getElementById("billingTxType").value,
  };

  try {
    await getBackend().CreateBillingCategory(payload);
    document.getElementById("billingCategoryForm").reset();
    await Promise.all([refreshBillingCategories(), refreshDashboard()]);
    showToast("Categoria de cobro creada.");
  } catch (error) {
    showToast(error.message || "No se pudo crear la categoria de cobro.", true);
  }
}

async function onBillingCategoriesAction(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;

  const { action, id } = button.dataset;
  const current = getBillingCategoryById(id);
  if (!current) return;

  try {
    if (action === "edit-billing") {
      const newName = prompt("Nuevo nombre:", current.name);
      if (newName === null) return;
      const newCalcType = prompt("Tipo de calculo (hourly o unit):", current.calc_type);
      if (newCalcType === null) return;
      const newRate = prompt("Tarifa base:", String(current.default_rate));
      if (newRate === null) return;
      const newUnitLabel = prompt("Etiqueta de unidad:", current.unit_label);
      if (newUnitLabel === null) return;
      const newTxType = prompt("Tipo de transaccion (income o expense):", current.transaction_type);
      if (newTxType === null) return;

      await getBackend().UpdateBillingCategory({
        id: current.id,
        name: newName.trim(),
        calc_type: newCalcType.trim(),
        default_rate: Number(newRate),
        unit_label: newUnitLabel.trim(),
        transaction_type: newTxType.trim(),
      });
      await Promise.all([refreshBillingCategories(), refreshDashboard()]);
      showToast("Categoria de cobro actualizada.");
    }

    if (action === "delete-billing") {
      const approved = confirm(`Eliminar la categoria de cobro '${current.name}'?`);
      if (!approved) return;

      await getBackend().DeleteBillingCategory(current.id);
      state.quoteResult = null;
      renderQuoteResult();
      await Promise.all([refreshBillingCategories(), refreshDashboard()]);
      showToast("Categoria de cobro eliminada.");
    }
  } catch (error) {
    showToast(error.message || "Operacion no completada.", true);
  }
}

async function onCalculateQuote(event) {
  event.preventDefault();

  const payload = {
    billing_category_id: quoteBillingCategoryEl.value,
    quantity: Number(document.getElementById("quoteQuantity").value),
    rate: Number(quoteRateEl.value),
    description: document.getElementById("quoteDescription").value,
  };

  try {
    state.quoteResult = await getBackend().CalculateQuote(payload);
    renderQuoteResult();
    showToast("Cotizacion calculada.");
  } catch (error) {
    showToast(error.message || "No se pudo calcular la cotizacion.", true);
  }
}

function openTransactionModal() {
  transactionModalEl.classList.remove("hidden");
  document.getElementById("transactionModalAmount").focus();
}

function closeTransactionModal() {
  transactionModalEl.classList.add("hidden");
  document.getElementById("transactionModalForm").reset();
  txModalState.categoryId = null;
}

function setTransactionTypeBadge(type) {
  const badge = document.getElementById("transactionModalTypeBadge");
  const isIncome = type === "income";
  badge.textContent = isIncome ? "Ingreso" : "Egreso";
  badge.classList.toggle("income", isIncome);
  badge.classList.toggle("expense", !isIncome);
}

function openTransactionModalForCategory(category) {
  txModalState.categoryId = category.id;
  txModalState.type = category.type;

  document.getElementById("transactionModalCategoryField").classList.add("hidden");
  const nameEl = document.getElementById("transactionModalCategoryName");
  nameEl.textContent = category.name;
  nameEl.classList.remove("hidden");
  setTransactionTypeBadge(category.type);

  document.getElementById("transactionModalAmount").value = "";
  document.getElementById("transactionModalDescription").value = "";
  openTransactionModal();
}

function openTransactionModalForType(type) {
  txModalState.categoryId = null;
  txModalState.type = type;

  document.getElementById("transactionModalCategoryField").classList.remove("hidden");
  document.getElementById("transactionModalCategoryName").classList.add("hidden");

  const categories = state.categories.filter((c) => c.type === type);
  transactionModalCategorySelectEl.innerHTML = categories
    .map((c) => `<option value='${c.id}'>${esc(c.name)}</option>`)
    .join("");

  setTransactionTypeBadge(type);

  document.getElementById("transactionModalAmount").value = "";
  document.getElementById("transactionModalDescription").value = "";
  openTransactionModal();
}

function useQuoteInTransaction() {
  if (!state.quoteResult) {
    showToast("Primero calcula una cotizacion.", true);
    return;
  }

  const type = state.quoteResult.suggested_type || "income";
  const categories = state.categories.filter((item) => item.type === type);
  if (categories.length === 0) {
    showToast("No hay categoria contable para este tipo. Crea una antes de guardar.", true);
    return;
  }

  openTransactionModalForType(type);
  document.getElementById("transactionModalAmount").value = Number(state.quoteResult.amount).toFixed(2);
  document.getElementById("transactionModalDescription").value = state.quoteResult.description || "";
}

async function onCreateTransactionFromModal(event) {
  event.preventDefault();

  let categoryId = txModalState.categoryId;
  let type = txModalState.type;

  if (!categoryId) {
    categoryId = transactionModalCategorySelectEl.value;
    const category = getCategoryById(categoryId);
    if (category) type = category.type;
  }

  const payload = {
    type,
    amount: Number(document.getElementById("transactionModalAmount").value),
    description: document.getElementById("transactionModalDescription").value,
    category_id: categoryId,
  };

  try {
    await getBackend().CreateTransaction(payload);
    closeTransactionModal();
    await Promise.all([refreshTransactions(), refreshDashboard(), refreshCategoryRanking()]);
    showToast("Transaccion guardada.");
  } catch (error) {
    showToast(error.message || "No se pudo guardar la transaccion.", true);
  }
}

async function onApplyFilters(event) {
  event.preventDefault();
  state.transactionFilters = readTransactionFiltersFromUI();

  try {
    await refreshTransactions();
    showToast("Filtros aplicados.");
  } catch (error) {
    showToast(error.message || "No se pudieron aplicar filtros.", true);
  }
}

async function onResetFilters() {
  state.transactionFilters = { start_date: "", end_date: "", category_id: "", type: "" };
  filterStartDateEl.value = "";
  filterEndDateEl.value = "";
  filterCategoryEl.value = "";
  filterTypeEl.value = "";

  try {
    await refreshTransactions();
    showToast("Filtros limpiados.");
  } catch (error) {
    showToast(error.message || "No se pudieron limpiar filtros.", true);
  }
}

function downloadCSV(content, filename) {
  const blob = new Blob(["\uFEFF" + content], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
}

async function onExportCSV() {
  state.transactionFilters = readTransactionFiltersFromUI();

  try {
    const csv = await getBackend().ExportTransactionsCSV(state.transactionFilters);
    if (!csv || csv.trim() === "") {
      showToast("No hay datos para exportar.", true);
      return;
    }

    const stamp = new Date().toISOString().slice(0, 19).replaceAll(":", "-");
    downloadCSV(csv, `transacciones_${stamp}.csv`);
    showToast("CSV exportado.");
  } catch (error) {
    showToast(error.message || "No se pudo exportar CSV.", true);
  }
}

async function onProviderSelectionChanged() {
  renderAIModelOptions(aiProviderSelectEl.value);
}

async function onSaveAIConfiguration(event) {
  event.preventDefault();
  try {
    const providerID = aiProviderSelectEl.value;
    const modelID = aiModelSelectEl.value;
    await getBackend().SaveAIConfiguration({
      provider_id: providerID,
      model_id: modelID,
      api_key: aiApiKeyEl.value.trim(),
    });
    aiApiKeyEl.value = "";
    await Promise.all([refreshAIProviders(), refreshAIModels()]);
    await refreshAIConfiguration();
    showToast("Configuracion IA guardada.");
  } catch (error) {
    showToast(error.message || "No se pudo guardar configuracion IA.", true);
  }
}

async function onCreateAIProvider(event) {
  event.preventDefault();
  const name = document.getElementById("newProviderName").value.trim();
  const baseURL = document.getElementById("newProviderURL").value.trim();
  const chatPath = document.getElementById("newProviderPath").value.trim();

  try {
    await getBackend().CreateAIProvider({
      name,
      base_url: baseURL,
      chat_path: chatPath || "/chat/completions",
    });
    document.getElementById("newProviderForm").reset();
    await Promise.all([refreshAIProviders(), refreshAIModels()]);
    await refreshAIConfiguration();
    showToast("Proveedor IA agregado.");
  } catch (error) {
    showToast(error.message || "No se pudo agregar proveedor.", true);
  }
}

async function onCreateAIModel(event) {
  event.preventDefault();
  const providerID = newModelProviderEl.value;
  const name = document.getElementById("newModelName").value.trim();

  try {
    await getBackend().CreateAIModel({
      provider_id: providerID,
      name,
    });
    document.getElementById("newModelForm").reset();
    await refreshAIModels();
    await refreshAIConfiguration();
    showToast("Modelo IA agregado.");
  } catch (error) {
    showToast(error.message || "No se pudo agregar modelo.", true);
  }
}

async function onAnalyzeDashboardWithAI(event) {
  event.preventDefault();
  const question = document.getElementById("aiQuestion").value.trim();

  try {
    const result = await getBackend().AnalyzeDashboardWithAI({ question });
    state.aiAnalysis = result;
    document.getElementById("aiQuestion").value = "";
    renderAIAnalysis();
    await refreshAIAnalyses();
    showToast("Analisis IA completado.");
  } catch (error) {
    showToast(error.message || "No se pudo generar analisis IA.", true);
  }
}

async function onAIAnalysesAction(event) {
  const button = event.target.closest("button[data-action]");
  if (!button) return;

  const { action, id } = button.dataset;

  try {
    if (action === "view-analysis") {
      const analysis = state.aiAnalyses.find((a) => a.id === id);
      if (!analysis) return;
      state.aiAnalysis = analysis;
      renderAIAnalysis();
      document.getElementById("aiAnalysisBox").scrollIntoView({ behavior: "smooth", block: "center" });
    }

    if (action === "delete-analysis") {
      const approved = confirm("Eliminar este analisis guardado?");
      if (!approved) return;

      await getBackend().DeleteAIAnalysis(id);
      if (state.aiAnalysis && state.aiAnalysis.id === id) {
        state.aiAnalysis = null;
        renderAIAnalysis();
      }
      await refreshAIAnalyses();
      showToast("Analisis eliminado.");
    }
  } catch (error) {
    showToast(error.message || "Operacion no completada.", true);
  }
}

async function onRankingMetricChange(metric) {
  if (metric !== "income" && metric !== "expense" && metric !== "mixed") return;
  state.rankingMetric = metric;

  document.querySelectorAll(".ranking-btn").forEach((btn) => {
    btn.classList.toggle("active", btn.dataset.metric === metric);
  });

  try {
    await refreshCategoryRanking();
  } catch (error) {
    showToast(error.message || "No se pudo actualizar el top de categorias.", true);
  }
}

async function onSaveAppSettings(event) {
  event.preventDefault();
  const days = Number(document.getElementById("appSettingsDays").value) || 0;

  try {
    await getBackend().SaveAppSettings({ ranking_days: days });
    state.appSettings = { ranking_days: days };
    await refreshCategoryRanking();
    showToast("Configuracion guardada.");
  } catch (error) {
    showToast(error.message || "No se pudo guardar la configuracion.", true);
  }
}

function wireEvents() {
  const menuBtn = document.getElementById("menuBtn");
  const menuDropdown = document.getElementById("menuDropdown");

  menuBtn.addEventListener("click", (event) => {
    event.stopPropagation();
    toggleMenu();
  });

  menuDropdown.addEventListener("click", (event) => {
    const tile = event.target.closest(".menu-tile");
    if (!tile) return;
    switchView(tile.dataset.view);
  });

  document.addEventListener("click", (event) => {
    if (
      !menuDropdown.classList.contains("hidden") &&
      !menuDropdown.contains(event.target) &&
      !menuBtn.contains(event.target)
    ) {
      closeMenu();
    }
  });

  document.getElementById("openHelpBtn").addEventListener("click", () => switchView("ayuda"));

  document.querySelector(".ranking-toolbar").addEventListener("click", (event) => {
    const button = event.target.closest("button[data-metric]");
    if (!button) return;
    onRankingMetricChange(button.dataset.metric);
  });

  document.getElementById("openCategoryModalBtn").addEventListener("click", openCategoryModal);
  document.getElementById("categoryModalForm").addEventListener("submit", onCreateCategoryFromModal);
  document.getElementById("categoryModalCancel").addEventListener("click", closeCategoryModal);
  document.querySelector(".type-toggle").addEventListener("click", onCategoryTypeClick);
  categoryModalEl.addEventListener("click", (event) => {
    if (event.target === categoryModalEl) closeCategoryModal();
  });

  document.getElementById("categoryBars").addEventListener("click", (event) => {
    const card = event.target.closest(".cat-card");
    if (!card) return;
    const category = getCategoryById(card.dataset.id);
    if (!category) return;
    openTransactionModalForCategory(category);
  });

  document.getElementById("transactionModalForm").addEventListener("submit", onCreateTransactionFromModal);
  document.getElementById("transactionModalCancel").addEventListener("click", closeTransactionModal);
  transactionModalEl.addEventListener("click", (event) => {
    if (event.target === transactionModalEl) closeTransactionModal();
  });

  document.getElementById("billingCategoryForm").addEventListener("submit", onCreateBillingCategory);
  document.getElementById("billingCategoriesList").addEventListener("click", onBillingCategoriesAction);

  document.getElementById("quoteForm").addEventListener("submit", onCalculateQuote);
  quoteBillingCategoryEl.addEventListener("change", () => {
    quoteRateEl.value = "";
    syncQuoteFieldsFromBillingCategory();
  });
  document.getElementById("useQuoteBtn").addEventListener("click", useQuoteInTransaction);

  document.getElementById("txFilterForm").addEventListener("submit", onApplyFilters);
  document.getElementById("resetFiltersBtn").addEventListener("click", onResetFilters);
  document.getElementById("exportCsvBtn").addEventListener("click", onExportCSV);

  document.getElementById("aiConfigForm").addEventListener("submit", onSaveAIConfiguration);
  aiProviderSelectEl.addEventListener("change", onProviderSelectionChanged);
  document.getElementById("newProviderForm").addEventListener("submit", onCreateAIProvider);
  document.getElementById("newModelForm").addEventListener("submit", onCreateAIModel);
  document.getElementById("appSettingsForm").addEventListener("submit", onSaveAppSettings);
  document.getElementById("aiAnalyzeForm").addEventListener("submit", onAnalyzeDashboardWithAI);
  document.getElementById("aiAnalysesList").addEventListener("click", onAIAnalysesAction);
}

(async function init() {
  wireEvents();
  switchView("dashboard");
  try {
    await refreshAll();
  } catch (error) {
    showToast(error.message || "Error al cargar datos iniciales.", true);
  }
})();
