const navButtons = Array.from(document.querySelectorAll(".nav-item"));
const sections = Array.from(document.querySelectorAll(".page-section"));
const messageBox = document.getElementById("message-box");
const sessionStatusTextEl = document.getElementById("session-status-text");
const btnOpenChangePassword = document.getElementById("btn-open-change-password");
const btnLogout = document.getElementById("btn-logout");

const btnCheckHealth = document.getElementById("btn-check-health");
const btnLoadConfig = document.getElementById("btn-load-config");
const btnLoadConfigInline = document.getElementById("btn-load-config-inline");

const btnRefreshLists = document.getElementById("btn-refresh-lists");
const btnResetListForm = document.getElementById("btn-reset-list-form");

const btnRefreshRules = document.getElementById("btn-refresh-rules");
const btnResetRuleForm = document.getElementById("btn-reset-rule-form");

const btnRefreshDesiredSources = document.getElementById("btn-refresh-desired-sources");
const btnResetDesiredSourceForm = document.getElementById("btn-reset-desired-source-form");
const btnRefreshCurrentSources = document.getElementById("btn-refresh-current-sources");
const btnResetCurrentSourceForm = document.getElementById("btn-reset-current-source-form");

const btnTestDesiredSource = document.getElementById("btn-test-desired-source");
const btnTestCurrentSource = document.getElementById("btn-test-current-source");
const btnSaveProbedSource = document.getElementById("btn-save-probed-source");
const btnClearSourceProbeResult = document.getElementById("btn-clear-source-probe-result");

const btnResetRenderForm = document.getElementById("btn-reset-render-form");
const btnClearRenderResult = document.getElementById("btn-clear-render-result");
const btnCopyRenderScript = document.getElementById("btn-copy-render-script");

const healthStatusEl = document.getElementById("dashboard-health-status");
const listCountEl = document.getElementById("dashboard-list-count");
const ruleCountEl = document.getElementById("dashboard-rule-count");
const renderModeEl = document.getElementById("dashboard-render-mode");
const desiredSourceCountEl = document.getElementById("dashboard-desired-source-count");
const currentSourceCountEl = document.getElementById("dashboard-current-source-count");
const summaryListEl = document.getElementById("dashboard-summary-list");

const configViewerEl = document.getElementById("config-json-viewer");

const listsTableBody = document.getElementById("lists-table-body");
const rulesTableBody = document.getElementById("rules-table-body");
const desiredSourcesTableBody = document.getElementById("desired-sources-table-body");
const currentSourcesTableBody = document.getElementById("current-sources-table-body");

const listForm = document.getElementById("list-form");
const listNameInput = document.getElementById("list-name");
const listFamilySelect = document.getElementById("list-family");
const listEnabledInput = document.getElementById("list-enabled");
const listDescriptionInput = document.getElementById("list-description");

const descriptionForm = document.getElementById("description-form");
const descTargetNameInput = document.getElementById("desc-target-name");
const descTextInput = document.getElementById("desc-text");

const ruleForm = document.getElementById("rule-form");
const ruleIdInput = document.getElementById("rule-id");
const ruleListNameInput = document.getElementById("rule-list-name");
const ruleActionSelect = document.getElementById("rule-action");
const rulePriorityInput = document.getElementById("rule-priority");
const ruleEnabledInput = document.getElementById("rule-enabled");
const ruleDescriptionInput = document.getElementById("rule-description");
const ruleEntriesInput = document.getElementById("rule-entries");

const desiredSourceForm = document.getElementById("desired-source-form");
const desiredSourceNameInput = document.getElementById("desired-source-name");
const desiredSourceTypeSelect = document.getElementById("desired-source-type");
const desiredSourcePathInput = document.getElementById("desired-source-path");
const desiredSourceURLInput = document.getElementById("desired-source-url");
const desiredSourceHeadersInput = document.getElementById("desired-source-headers");
const desiredSourcePriorityInput = document.getElementById("desired-source-priority");
const desiredSourceTimeoutInput = document.getElementById("desired-source-timeout");
const desiredSourceEnabledInput = document.getElementById("desired-source-enabled");
const desiredSourcePathHelp = document.getElementById("desired-source-path-help");
const desiredSourceURLHelp = document.getElementById("desired-source-url-help");

const currentSourceForm = document.getElementById("current-source-form");
const currentSourceNameInput = document.getElementById("current-source-name");
const currentSourceTypeSelect = document.getElementById("current-source-type");
const currentSourcePathInput = document.getElementById("current-source-path");
const currentSourceURLInput = document.getElementById("current-source-url");
const currentSourceHeadersInput = document.getElementById("current-source-headers");
const currentSourcePriorityInput = document.getElementById("current-source-priority");
const currentSourceTimeoutInput = document.getElementById("current-source-timeout");
const currentSourceEnabledInput = document.getElementById("current-source-enabled");
const currentSourcePathHelp = document.getElementById("current-source-path-help");
const currentSourceURLHelp = document.getElementById("current-source-url-help");

// Source Probe 基础字段
const sourceProbeStatusEl = document.getElementById("source-probe-status");
const sourceProbeNameEl = document.getElementById("source-probe-name");
const sourceProbeLocationEl = document.getElementById("source-probe-location");
const sourceProbeHTTPStatusEl = document.getElementById("source-probe-http-status");
const sourceProbeContentTypeEl = document.getElementById("source-probe-content-type");
const sourceProbeBodyBytesEl = document.getElementById("source-probe-body-bytes");
const sourceProbeJSONValidEl = document.getElementById("source-probe-json-valid");
const sourceProbeJSONTypeEl = document.getElementById("source-probe-json-type");
const sourceProbeListCountEl = document.getElementById("source-probe-list-count");
const sourceProbeEntryCountEl = document.getElementById("source-probe-entry-count");
const sourceProbeHeadersCountEl = document.getElementById("source-probe-headers-count");
const sourceProbeDurationEl = document.getElementById("source-probe-duration");
const sourceProbeErrorEl = document.getElementById("source-probe-error");
const sourceProbeListNamesEl = document.getElementById("source-probe-list-names");
const sourceProbeResponseHeadersEl = document.getElementById("source-probe-response-headers");
const sourceProbeRawPreviewEl = document.getElementById("source-probe-raw-preview");

// Source Probe 新增字段（如果 index.html 尚未替换，这些可能为 null，不影响运行）
const sourceProbeDetectedFormatEl = document.getElementById("source-probe-detected-format");
const sourceProbeFormatDetailEl = document.getElementById("source-probe-format-detail");
const sourceProbeTextLineCountEl = document.getElementById("source-probe-text-line-count");
const sourceProbeTextValidLineCountEl = document.getElementById("source-probe-text-valid-line-count");
const sourceProbeTextInvalidLineCountEl = document.getElementById("source-probe-text-invalid-line-count");

const renderForm = document.getElementById("render-form");
const renderModeSelect = document.getElementById("render-mode");
const renderOutputPathInput = document.getElementById("render-output-path");
const renderResultModeEl = document.getElementById("render-result-mode");
const renderResultListCountEl = document.getElementById("render-result-list-count");
const renderResultEntryCountEl = document.getElementById("render-result-entry-count");
const renderResultOutputPathEl = document.getElementById("render-result-output-path");
const renderScriptViewerEl = document.getElementById("render-script-viewer");

let currentConfig = null;
let currentSessionProfile = null;
let currentLists = [];
let currentRules = [];
let currentDesiredSources = [];
let currentCurrentSources = [];

let editingListName = "";
let editingRuleId = "";
let editingDesiredSourceName = "";
let editingCurrentSourceName = "";

let currentRenderResult = null;

const pageQuery = new URLSearchParams(window.location.search);

let lastProbeKind = "";
let lastProbePayload = null;
let lastProbeSourceName = "";
let lastProbeResult = null;

function escapeHTML(value) {
    return String(value ?? "")
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#39;");
}

function showMessage(message, isError = false) {
    if (!messageBox) return;
    messageBox.textContent = String(message || "");
    messageBox.classList.remove("hidden");
    messageBox.classList.toggle("error", Boolean(isError));
}

function clearMessage() {
    if (!messageBox) return;
    messageBox.textContent = "";
    messageBox.classList.add("hidden");
    messageBox.classList.remove("error");
}

function setActiveSection(sectionId) {
    clearMessage();

    navButtons.forEach((button) => {
        const active = button.dataset.section === sectionId;
        button.classList.toggle("active", active);
    });

    sections.forEach((section) => {
        const active = section.id === sectionId;
        section.classList.toggle("active", active);
    });
}

function setCompactValue(el, value) {
    if (!el) return;
    if (value === null || value === undefined || value === "") {
        el.textContent = "-";
        return;
    }
    el.textContent = String(value);
}

function setPreBlock(el, value, emptyText = "-") {
    if (!el) return;
    const text = String(value ?? "").trim();
    el.textContent = text || emptyText;
    el.classList.toggle("empty-viewer", !text);
}

function safeJSONStringify(value) {
    try {
        return JSON.stringify(value ?? null, null, 2);
    } catch (_) {
        return String(value ?? "");
    }
}

async function apiFetch(path, options = {}) {
    const response = await fetch(path, {
        credentials: "include",
        headers: {
            "Content-Type": "application/json; charset=utf-8",
            ...(options.headers || {}),
        },
        ...options,
    });

    const text = await response.text();
    let data = null;

    try {
        data = text ? JSON.parse(text) : null;
    } catch (_) {
        throw new Error(`响应不是合法 JSON：${text}`);
    }

    if (!response.ok) {
        const message =
            data && typeof data.error === "string"
                ? data.error
                : `HTTP ${response.status}`;

        if (response.status === 401 && path !== "/api/v1/auth/me") {
            window.location.href = "/login.html";
        }

        if (
            response.status === 403 &&
            data &&
            (data.must_change_password === true || message.includes("password change required"))
        ) {
            window.location.href = "/login.html?force_change=1";
        }

        throw new Error(message);
    }

    return data;
}

async function ensureAuthenticatedPage() {
    try {
        const profile = await apiFetch("/api/v1/auth/me");
        currentSessionProfile = profile;

        if (profile && profile.must_change_password) {
            window.location.href = "/login.html?force_change=1";
            throw new Error("首次登录需要先修改密码");
        }

        renderSessionPanel(profile);
        return profile;
    } catch (error) {
        if (!String(error?.message || "").includes("首次登录需要先修改密码")) {
            window.location.href = "/login.html";
        }
        throw error;
    }
}

function renderSessionPanel(profile) {
    if (!sessionStatusTextEl) return;

    if (!profile || !profile.authenticated) {
        sessionStatusTextEl.textContent = "未登录";
        return;
    }

    sessionStatusTextEl.textContent = `当前已登录：${profile.username || "admin"}。登录后才能访问后台各项设置。`;
}

function parseHeadersText(text) {
    const lines = String(text || "")
        .split("\n")
        .map((line) => line.trim())
        .filter(Boolean);

    const headers = {};

    for (const line of lines) {
        const index = line.indexOf(":");
        if (index <= 0) {
            throw new Error(`Headers 格式错误：${line}`);
        }
        const key = line.slice(0, index).trim();
        const value = line.slice(index + 1).trim();
        if (!key) {
            throw new Error(`Headers key 不能为空：${line}`);
        }
        headers[key] = value;
    }

    return headers;
}

function stringifyHeadersMap(headers) {
    if (!headers || typeof headers !== "object") {
        return "";
    }

    return Object.entries(headers)
        .map(([key, value]) => `${key}: ${value}`)
        .join("\n");
}

function makeActionButton(label, className, onClick) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = className || "btn btn-sm";
    button.textContent = label;
    button.addEventListener("click", onClick);
    return button;
}

function appendEmptyRow(tbody, colSpan, text) {
    tbody.innerHTML = "";
    const tr = document.createElement("tr");
    const td = document.createElement("td");
    td.colSpan = colSpan;
    td.textContent = text;
    td.style.textAlign = "center";
    td.style.opacity = "0.75";
    tr.appendChild(td);
    tbody.appendChild(tr);
}

function boolText(v) {
    return v ? "true" : "false";
}

function getSourceLocation(item) {
    return item?.type === "url" ? (item?.url || "") : (item?.path || "");
}

function syncSourceTypeUI(kind) {
    const isDesired = kind === "desired";

    const typeSelect = isDesired ? desiredSourceTypeSelect : currentSourceTypeSelect;
    const pathInput = isDesired ? desiredSourcePathInput : currentSourcePathInput;
    const urlInput = isDesired ? desiredSourceURLInput : currentSourceURLInput;
    const headersInput = isDesired ? desiredSourceHeadersInput : currentSourceHeadersInput;
    const pathHelp = isDesired ? desiredSourcePathHelp : currentSourcePathHelp;
    const urlHelp = isDesired ? desiredSourceURLHelp : currentSourceURLHelp;

    if (!typeSelect) return;

    const isFile = typeSelect.value === "file";
    const isURL = typeSelect.value === "url";

    if (pathInput) pathInput.disabled = !isFile;
    if (urlInput) urlInput.disabled = !isURL;
    if (headersInput) headersInput.disabled = !isURL;

    if (pathHelp) {
        pathHelp.style.opacity = isFile ? "1" : "0.55";
    }
    if (urlHelp) {
        urlHelp.style.opacity = isURL ? "1" : "0.55";
    }
}

function renderConfigToDashboard(config) {
    if (!config) return;

    setCompactValue(listCountEl, config.lists?.length ?? 0);
    setCompactValue(ruleCountEl, config.manual_rules?.length ?? 0);
    setCompactValue(renderModeEl, config.output?.mode ?? "-");
    setCompactValue(desiredSourceCountEl, config.desired_sources?.length ?? 0);
    setCompactValue(currentSourceCountEl, config.current_state_sources?.length ?? 0);

    if (summaryListEl) {
        summaryListEl.innerHTML = "";

        const summaryItems = [
            `自动创建未知 list：${String(config.auto_create_lists)}`,
            `日志文件：${config.log_file ?? "-"}`,
            `输出路径：${config.output?.path ?? "-"}`,
            `managed comment：${config.output?.managed_comment ?? "-"}`,
            `监听地址：${config.server?.listen ?? "-"}`,
            `desired_sources 数量：${config.desired_sources?.length ?? 0}`,
            `current_state_sources 数量：${config.current_state_sources?.length ?? 0}`,
        ];

        summaryItems.forEach((item) => {
            const li = document.createElement("li");
            li.textContent = item;
            summaryListEl.appendChild(li);
        });
    }
}

function renderConfigToViewer(config) {
    if (!configViewerEl) return;
    configViewerEl.textContent = safeJSONStringify(config);
    configViewerEl.classList.toggle("empty-viewer", !config);
}

async function fetchHealthStatus() {
    setCompactValue(healthStatusEl, "检查中...");
    const data = await apiFetch("/healthz");
    setCompactValue(healthStatusEl, data?.status || "unknown");
    return data;
}

async function checkHealth() {
    clearMessage();
    try {
        await fetchHealthStatus();
        showMessage("服务状态检查成功。");
    } catch (error) {
        setCompactValue(healthStatusEl, "失败");
        showMessage(`服务状态检查失败：${error.message}`, true);
    }
}

async function loadConfig() {
    try {
        const data = await apiFetch("/api/v1/config");
        currentConfig = data;
        renderConfigToDashboard(data);
        renderConfigToViewer(data);

        if (renderModeSelect && renderModeSelect.value === "") {
            renderModeSelect.value = data?.output?.mode || "";
        }

        if (renderOutputPathInput && !renderOutputPathInput.value) {
            renderOutputPathInput.value = data?.output?.path || "";
        }
    } catch (error) {
        showMessage(`配置加载失败：${error.message}`, true);
    }
}

// Lists
async function loadLists() {
    try {
        const lists = await apiFetch("/api/v1/lists");
        currentLists = Array.isArray(lists) ? lists : [];
        renderListsTable(currentLists);

        if (currentConfig) {
            currentConfig.lists = currentLists;
            renderConfigToDashboard(currentConfig);
            renderConfigToViewer(currentConfig);
        }
    } catch (error) {
        showMessage(`加载 Address Lists 失败：${error.message}`, true);
    }
}

function renderListsTable(lists) {
    if (!listsTableBody) return;

    if (!lists.length) {
        appendEmptyRow(listsTableBody, 5, "当前没有任何 list");
        return;
    }

    listsTableBody.innerHTML = "";

    lists.forEach((item) => {
        const tr = document.createElement("tr");

        tr.innerHTML = `
      <td>${escapeHTML(item.name)}</td>
      <td>${escapeHTML(item.family)}</td>
      <td>${boolText(item.enabled)}</td>
      <td>${escapeHTML(item.description || "")}</td>
      <td></td>
    `;

        const actionCell = tr.lastElementChild;

        actionCell.appendChild(
            makeActionButton("编辑", "btn btn-sm", () => editList(item.name))
        );
        actionCell.appendChild(
            makeActionButton("改说明", "btn btn-sm", () => fillDescriptionForm(item))
        );
        actionCell.appendChild(
            makeActionButton("删除", "btn btn-sm btn-danger", () => deleteList(item.name))
        );

        listsTableBody.appendChild(tr);
    });
}

function fillDescriptionForm(item) {
    if (!item) return;
    descTargetNameInput.value = item.name || "";
    descTextInput.value = item.description || "";
}

async function getListByName(name) {
    return apiFetch(`/api/v1/lists/${encodeURIComponent(name)}`);
}

async function editList(name) {
    clearMessage();
    try {
        const item = await getListByName(name);
        editingListName = item.name || "";
        listNameInput.value = item.name || "";
        listFamilySelect.value = item.family || "ipv4";
        listEnabledInput.checked = Boolean(item.enabled);
        listDescriptionInput.value = item.description || "";
        fillDescriptionForm(item);
        setActiveSection("lists-section");
        showMessage(`已加载 ${name} 到 list 编辑表单。`);
    } catch (error) {
        showMessage(`加载 list 详情失败：${error.message}`, true);
    }
}

async function deleteList(name) {
    const ok = window.confirm(`确定删除 list "${name}" 吗？`);
    if (!ok) return;

    clearMessage();

    try {
        await apiFetch(`/api/v1/lists/${encodeURIComponent(name)}`, {
            method: "DELETE",
        });

        if (editingListName === name) {
            resetListForm();
        }
        if (descTargetNameInput.value === name) {
            descTargetNameInput.value = "";
            descTextInput.value = "";
        }

        await loadLists();
        await loadConfig();

        showMessage(`已删除 list：${name}`);
    } catch (error) {
        showMessage(`删除 list 失败：${error.message}`, true);
    }
}

function resetListForm() {
    editingListName = "";
    if (listForm) listForm.reset();
    if (listFamilySelect) listFamilySelect.value = "ipv4";
    if (listEnabledInput) listEnabledInput.checked = true;
}

async function submitListForm(event) {
    event.preventDefault();
    clearMessage();

    const name = listNameInput.value.trim();
    const family = listFamilySelect.value;
    const enabled = Boolean(listEnabledInput.checked);
    const description = listDescriptionInput.value;

    if (!name) {
        showMessage("List Name 不能为空。", true);
        return;
    }

    const payload = { name, family, enabled, description };

    try {
        if (editingListName && editingListName === name) {
            await apiFetch(`/api/v1/lists/${encodeURIComponent(name)}`, {
                method: "PUT",
                body: JSON.stringify(payload),
            });
            showMessage(`已更新 list：${name}`);
        } else {
            await apiFetch("/api/v1/lists", {
                method: "POST",
                body: JSON.stringify(payload),
            });
            showMessage(`已新增 list：${name}`);
        }

        await loadLists();
        await loadConfig();
        editingListName = name;
    } catch (error) {
        showMessage(`保存 list 失败：${error.message}`, true);
    }
}

async function submitDescriptionForm(event) {
    event.preventDefault();
    clearMessage();

    const name = descTargetNameInput.value.trim();
    const description = descTextInput.value;

    if (!name) {
        showMessage("目标 Name 不能为空。", true);
        return;
    }

    try {
        await apiFetch(`/api/v1/lists/${encodeURIComponent(name)}/description`, {
            method: "PUT",
            body: JSON.stringify({ description }),
        });

        await loadLists();
        await loadConfig();

        showMessage(`已更新 ${name} 的 description。`);
    } catch (error) {
        showMessage(`更新 description 失败：${error.message}`, true);
    }
}

// Rules
async function loadRules() {
    try {
        const rules = await apiFetch("/api/v1/manual-rules");
        currentRules = Array.isArray(rules) ? rules : [];
        renderRulesTable(currentRules);

        if (currentConfig) {
            currentConfig.manual_rules = currentRules;
            renderConfigToDashboard(currentConfig);
            renderConfigToViewer(currentConfig);
        }
    } catch (error) {
        showMessage(`加载 Manual Rules 失败：${error.message}`, true);
    }
}

function renderRulesTable(rules) {
    if (!rulesTableBody) return;

    if (!rules.length) {
        appendEmptyRow(rulesTableBody, 8, "当前没有任何 rule");
        return;
    }

    rulesTableBody.innerHTML = "";

    rules.forEach((item) => {
        const tr = document.createElement("tr");
        const entriesText = Array.isArray(item.entries) ? item.entries.join(", ") : "";

        tr.innerHTML = `
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.list_name)}</td>
      <td>${escapeHTML(item.action)}</td>
      <td>${escapeHTML(String(item.priority ?? ""))}</td>
      <td>${boolText(item.enabled)}</td>
      <td>${escapeHTML(item.description || "")}</td>
      <td>${escapeHTML(entriesText)}</td>
      <td></td>
    `;

        const actionCell = tr.lastElementChild;

        actionCell.appendChild(
            makeActionButton("编辑", "btn btn-sm", () => editRule(item.id))
        );
        actionCell.appendChild(
            makeActionButton("删除", "btn btn-sm btn-danger", () => deleteRule(item.id))
        );

        rulesTableBody.appendChild(tr);
    });
}

function resetRuleForm() {
    editingRuleId = "";
    if (ruleForm) ruleForm.reset();
    if (ruleActionSelect) ruleActionSelect.value = "add";
    if (rulePriorityInput) rulePriorityInput.value = "1000";
    if (ruleEnabledInput) ruleEnabledInput.checked = true;
}

function fillRuleForm(rule) {
    if (!rule) return;

    editingRuleId = rule.id || "";
    ruleIdInput.value = rule.id || "";
    ruleListNameInput.value = rule.list_name || "";
    ruleActionSelect.value = rule.action || "add";
    rulePriorityInput.value = String(rule.priority ?? 1000);
    ruleEnabledInput.checked = Boolean(rule.enabled);
    ruleDescriptionInput.value = rule.description || "";
    ruleEntriesInput.value = Array.isArray(rule.entries)
        ? rule.entries.join("\n")
        : "";
}

function findRuleById(id) {
    return currentRules.find((item) => item.id === id) || null;
}

async function editRule(id) {
    const rule = findRuleById(id);
    if (!rule) {
        showMessage(`未找到 rule：${id}`, true);
        return;
    }

    fillRuleForm(rule);
    setActiveSection("rules-section");
    showMessage(`已加载 ${id} 到 rule 编辑表单。`);
}

async function deleteRule(id) {
    const ok = window.confirm(`确定删除 rule "${id}" 吗？`);
    if (!ok) return;

    clearMessage();

    try {
        await apiFetch(`/api/v1/manual-rules/${encodeURIComponent(id)}`, {
            method: "DELETE",
        });

        if (editingRuleId === id) {
            resetRuleForm();
        }

        await loadRules();
        await loadConfig();

        showMessage(`已删除 rule：${id}`);
    } catch (error) {
        showMessage(`删除 rule 失败：${error.message}`, true);
    }
}

async function submitRuleForm(event) {
    event.preventDefault();
    clearMessage();

    const id = ruleIdInput.value.trim();
    const listName = ruleListNameInput.value.trim();
    const action = ruleActionSelect.value;
    const priority = Number(rulePriorityInput.value || 0);
    const enabled = Boolean(ruleEnabledInput.checked);
    const description = ruleDescriptionInput.value;
    const entries = ruleEntriesInput.value
        .split("\n")
        .map((item) => item.trim())
        .filter(Boolean);

    if (!listName) {
        showMessage("Rule 的 List Name 不能为空。", true);
        return;
    }

    if (!Number.isFinite(priority)) {
        showMessage("Priority 必须是数字。", true);
        return;
    }

    const payload = {
        id,
        list_name: listName,
        action,
        priority,
        enabled,
        description,
        entries,
    };

    try {
        if (editingRuleId) {
            await apiFetch(`/api/v1/manual-rules/${encodeURIComponent(editingRuleId)}`, {
                method: "PUT",
                body: JSON.stringify(payload),
            });
            showMessage(`已更新 rule：${editingRuleId}`);
        } else {
            await apiFetch("/api/v1/manual-rules", {
                method: "POST",
                body: JSON.stringify(payload),
            });
            showMessage(`已新增 rule：${id || "（由后端生成 ID）"}`);
        }

        await loadRules();
        await loadConfig();

        if (!editingRuleId && id) {
            editingRuleId = id;
        }
    } catch (error) {
        showMessage(`保存 rule 失败：${error.message}`, true);
    }
}

// Sources
function resetDesiredSourceForm() {
    editingDesiredSourceName = "";
    if (desiredSourceForm) desiredSourceForm.reset();
    desiredSourceHeadersInput.value = "";
    desiredSourceTypeSelect.value = "file";
    desiredSourcePriorityInput.value = "100";
    desiredSourceTimeoutInput.value = "15";
    desiredSourceEnabledInput.checked = true;
    syncSourceTypeUI("desired");
}

function resetCurrentSourceForm() {
    editingCurrentSourceName = "";
    if (currentSourceForm) currentSourceForm.reset();
    currentSourceHeadersInput.value = "";
    currentSourceTypeSelect.value = "file";
    currentSourcePriorityInput.value = "100";
    currentSourceTimeoutInput.value = "15";
    currentSourceEnabledInput.checked = true;
    syncSourceTypeUI("current");
}

async function loadDesiredSources() {
    try {
        const items = await apiFetch("/api/v1/sources/desired");
        currentDesiredSources = Array.isArray(items) ? items : [];
        renderSourceTable(desiredSourcesTableBody, currentDesiredSources, "desired");

        if (currentConfig) {
            currentConfig.desired_sources = currentDesiredSources;
            renderConfigToDashboard(currentConfig);
            renderConfigToViewer(currentConfig);
        }
    } catch (error) {
        showMessage(`加载 Desired Sources 失败：${error.message}`, true);
    }
}

async function loadCurrentSources() {
    try {
        const items = await apiFetch("/api/v1/sources/current");
        currentCurrentSources = Array.isArray(items) ? items : [];
        renderSourceTable(currentSourcesTableBody, currentCurrentSources, "current");

        if (currentConfig) {
            currentConfig.current_state_sources = currentCurrentSources;
            renderConfigToDashboard(currentConfig);
            renderConfigToViewer(currentConfig);
        }
    } catch (error) {
        showMessage(`加载 Current Sources 失败：${error.message}`, true);
    }
}

function renderSourceTable(targetBody, items, kind) {
    if (!targetBody) return;

    if (!items.length) {
        appendEmptyRow(targetBody, 8, "当前没有任何 source");
        return;
    }

    targetBody.innerHTML = "";

    items.forEach((item) => {
        const tr = document.createElement("tr");
        const location = getSourceLocation(item);
        const headerCount = item.headers ? Object.keys(item.headers).length : 0;
        const headerText = headerCount > 0 ? `${headerCount} 个` : "-";
        const enabledText = boolText(item.enabled);
        const toggleText = item.enabled ? "停用" : "启用";

        tr.innerHTML = `
      <td>${escapeHTML(item.name)}</td>
      <td>${escapeHTML(item.type)}</td>
      <td>${escapeHTML(location)}</td>
      <td>${escapeHTML(headerText)}</td>
      <td>${enabledText}</td>
      <td>${escapeHTML(String(item.priority ?? ""))}</td>
      <td>${escapeHTML(String(item.timeout_seconds ?? ""))}</td>
      <td></td>
    `;

        const actionCell = tr.lastElementChild;

        actionCell.appendChild(
            makeActionButton("编辑", "btn btn-sm", () => editSource(kind, item.name))
        );
        actionCell.appendChild(
            makeActionButton(toggleText, "btn btn-sm", () => toggleSourceEnabled(kind, item.name))
        );
        actionCell.appendChild(
            makeActionButton("删除", "btn btn-sm btn-danger", () => deleteSource(kind, item.name))
        );

        targetBody.appendChild(tr);
    });
}

function fillDesiredSourceForm(item) {
    editingDesiredSourceName = item?.name || "";
    desiredSourceNameInput.value = item?.name || "";
    desiredSourceTypeSelect.value = item?.type || "file";
    desiredSourcePathInput.value = item?.path || "";
    desiredSourceURLInput.value = item?.url || "";
    desiredSourceHeadersInput.value = stringifyHeadersMap(item?.headers);
    desiredSourcePriorityInput.value = String(item?.priority ?? 100);
    desiredSourceTimeoutInput.value = String(item?.timeout_seconds ?? 15);
    desiredSourceEnabledInput.checked = Boolean(item?.enabled);
    syncSourceTypeUI("desired");
}

function fillCurrentSourceForm(item) {
    editingCurrentSourceName = item?.name || "";
    currentSourceNameInput.value = item?.name || "";
    currentSourceTypeSelect.value = item?.type || "file";
    currentSourcePathInput.value = item?.path || "";
    currentSourceURLInput.value = item?.url || "";
    currentSourceHeadersInput.value = stringifyHeadersMap(item?.headers);
    currentSourcePriorityInput.value = String(item?.priority ?? 100);
    currentSourceTimeoutInput.value = String(item?.timeout_seconds ?? 15);
    currentSourceEnabledInput.checked = Boolean(item?.enabled);
    syncSourceTypeUI("current");
}

function findSource(kind, name) {
    const list = kind === "desired" ? currentDesiredSources : currentCurrentSources;
    return list.find((item) => item.name === name) || null;
}

async function getSourceByName(kind, name) {
    const path =
        kind === "desired"
            ? `/api/v1/sources/desired/${encodeURIComponent(name)}`
            : `/api/v1/sources/current/${encodeURIComponent(name)}`;

    return apiFetch(path);
}

async function editSource(kind, name) {
    clearMessage();

    try {
        let item = findSource(kind, name);
        if (!item) {
            item = await getSourceByName(kind, name);
        }

        if (!item) {
            showMessage(`未找到 source：${name}`, true);
            return;
        }

        if (kind === "desired") {
            fillDesiredSourceForm(item);
        } else {
            fillCurrentSourceForm(item);
        }

        setActiveSection("sources-section");
        showMessage(`已加载 ${name} 到 ${kind} source 编辑表单。`);
    } catch (error) {
        showMessage(`加载 source 详情失败：${error.message}`, true);
    }
}

async function toggleSourceEnabled(kind, name) {
    const item = findSource(kind, name);
    if (!item) {
        showMessage(`未找到 source：${name}`, true);
        return;
    }

    const payload = {
        ...item,
        enabled: !item.enabled,
    };

    const path =
        kind === "desired"
            ? `/api/v1/sources/desired/${encodeURIComponent(name)}`
            : `/api/v1/sources/current/${encodeURIComponent(name)}`;

    clearMessage();

    try {
        await apiFetch(path, {
            method: "PUT",
            body: JSON.stringify(payload),
        });

        if (kind === "desired") {
            await loadDesiredSources();
        } else {
            await loadCurrentSources();
        }
        await loadConfig();

        showMessage(
            `已${payload.enabled ? "启用" : "停用"} source：${name}`
        );
    } catch (error) {
        showMessage(`切换 source 状态失败：${error.message}`, true);
    }
}

async function deleteSource(kind, name) {
    const ok = window.confirm(`确定删除 source "${name}" 吗？`);
    if (!ok) return;

    const path =
        kind === "desired"
            ? `/api/v1/sources/desired/${encodeURIComponent(name)}`
            : `/api/v1/sources/current/${encodeURIComponent(name)}`;

    clearMessage();

    try {
        await apiFetch(path, {
            method: "DELETE",
        });

        if (kind === "desired") {
            if (editingDesiredSourceName === name) resetDesiredSourceForm();
            await loadDesiredSources();
        } else {
            if (editingCurrentSourceName === name) resetCurrentSourceForm();
            await loadCurrentSources();
        }

        await loadConfig();
        showMessage(`已删除 source：${name}`);
    } catch (error) {
        showMessage(`删除 source 失败：${error.message}`, true);
    }
}

function buildSourcePayloadFromDesiredForm() {
    return {
        name: desiredSourceNameInput.value.trim(),
        type: desiredSourceTypeSelect.value,
        path: desiredSourcePathInput.value.trim(),
        url: desiredSourceURLInput.value.trim(),
        headers: parseHeadersText(desiredSourceHeadersInput.value),
        priority: Number(desiredSourcePriorityInput.value || 0),
        timeout_seconds: Number(desiredSourceTimeoutInput.value || 15),
        enabled: Boolean(desiredSourceEnabledInput.checked),
    };
}

function buildSourcePayloadFromCurrentForm() {
    return {
        name: currentSourceNameInput.value.trim(),
        type: currentSourceTypeSelect.value,
        path: currentSourcePathInput.value.trim(),
        url: currentSourceURLInput.value.trim(),
        headers: parseHeadersText(currentSourceHeadersInput.value),
        priority: Number(currentSourcePriorityInput.value || 0),
        timeout_seconds: Number(currentSourceTimeoutInput.value || 15),
        enabled: Boolean(currentSourceEnabledInput.checked),
    };
}

async function submitDesiredSourceForm(event) {
    event.preventDefault();
    clearMessage();

    let payload;
    try {
        payload = buildSourcePayloadFromDesiredForm();
    } catch (error) {
        showMessage(`Desired Source Headers 解析失败：${error.message}`, true);
        return;
    }

    if (!payload.name) {
        showMessage("Desired Source Name 不能为空。", true);
        return;
    }

    if (payload.type === "file" && !payload.path) {
        showMessage("type=file 时 Path 不能为空。", true);
        return;
    }

    if (payload.type === "url" && !payload.url) {
        showMessage("type=url 时 URL 不能为空。", true);
        return;
    }

    if (!Number.isFinite(payload.priority)) {
        showMessage("Priority 必须是数字。", true);
        return;
    }

    if (!Number.isFinite(payload.timeout_seconds) || payload.timeout_seconds <= 0) {
        showMessage("Timeout Seconds 必须是大于 0 的数字。", true);
        return;
    }

    try {
        if (editingDesiredSourceName) {
            await apiFetch(`/api/v1/sources/desired/${encodeURIComponent(editingDesiredSourceName)}`, {
                method: "PUT",
                body: JSON.stringify(payload),
            });
            showMessage(`已更新 Desired Source：${editingDesiredSourceName}`);
        } else {
            await apiFetch("/api/v1/sources/desired", {
                method: "POST",
                body: JSON.stringify(payload),
            });
            showMessage(`已新增 Desired Source：${payload.name}`);
        }

        await loadDesiredSources();
        await loadConfig();
        editingDesiredSourceName = payload.name;
    } catch (error) {
        showMessage(`保存 Desired Source 失败：${error.message}`, true);
    }
}

async function submitCurrentSourceForm(event) {
    event.preventDefault();
    clearMessage();

    let payload;
    try {
        payload = buildSourcePayloadFromCurrentForm();
    } catch (error) {
        showMessage(`Current Source Headers 解析失败：${error.message}`, true);
        return;
    }

    if (!payload.name) {
        showMessage("Current Source Name 不能为空。", true);
        return;
    }

    if (payload.type === "file" && !payload.path) {
        showMessage("type=file 时 Path 不能为空。", true);
        return;
    }

    if (payload.type === "url" && !payload.url) {
        showMessage("type=url 时 URL 不能为空。", true);
        return;
    }

    if (!Number.isFinite(payload.priority)) {
        showMessage("Priority 必须是数字。", true);
        return;
    }

    if (!Number.isFinite(payload.timeout_seconds) || payload.timeout_seconds <= 0) {
        showMessage("Timeout Seconds 必须是大于 0 的数字。", true);
        return;
    }

    try {
        if (editingCurrentSourceName) {
            await apiFetch(`/api/v1/sources/current/${encodeURIComponent(editingCurrentSourceName)}`, {
                method: "PUT",
                body: JSON.stringify(payload),
            });
            showMessage(`已更新 Current Source：${editingCurrentSourceName}`);
        } else {
            await apiFetch("/api/v1/sources/current", {
                method: "POST",
                body: JSON.stringify(payload),
            });
            showMessage(`已新增 Current Source：${payload.name}`);
        }

        await loadCurrentSources();
        await loadConfig();
        editingCurrentSourceName = payload.name;
    } catch (error) {
        showMessage(`保存 Current Source 失败：${error.message}`, true);
    }
}

// Source Probe
function setSourceProbeLoading(kind, payload) {
    lastProbeKind = kind;
    lastProbePayload = payload;
    lastProbeSourceName = payload?.name || "";
    lastProbeResult = null;

    setCompactValue(sourceProbeStatusEl, "测试中...");
    setCompactValue(sourceProbeNameEl, payload?.name || "-");
    setCompactValue(
        sourceProbeLocationEl,
        payload?.type === "file" ? (payload?.path || "-") : (payload?.url || "-")
    );
    setCompactValue(sourceProbeHTTPStatusEl, "-");
    setCompactValue(sourceProbeContentTypeEl, "-");
    setCompactValue(sourceProbeBodyBytesEl, "-");
    setCompactValue(sourceProbeHeadersCountEl, "-");
    setCompactValue(sourceProbeDurationEl, "-");

    setCompactValue(sourceProbeDetectedFormatEl, "-");
    setCompactValue(sourceProbeFormatDetailEl, "-");

    setCompactValue(sourceProbeJSONValidEl, "-");
    setCompactValue(sourceProbeJSONTypeEl, "-");
    setCompactValue(sourceProbeListCountEl, "-");
    setCompactValue(sourceProbeEntryCountEl, "-");

    setCompactValue(sourceProbeTextLineCountEl, "-");
    setCompactValue(sourceProbeTextValidLineCountEl, "-");
    setCompactValue(sourceProbeTextInvalidLineCountEl, "-");

    setPreBlock(sourceProbeErrorEl, "测试进行中...");
    setPreBlock(sourceProbeListNamesEl, "");
    setPreBlock(sourceProbeResponseHeadersEl, "");
    setPreBlock(sourceProbeRawPreviewEl, "");
}

function clearSourceProbeResultView() {
    lastProbeKind = "";
    lastProbePayload = null;
    lastProbeSourceName = "";
    lastProbeResult = null;

    setCompactValue(sourceProbeStatusEl, "-");
    setCompactValue(sourceProbeNameEl, "-");
    setCompactValue(sourceProbeLocationEl, "-");
    setCompactValue(sourceProbeHTTPStatusEl, "-");
    setCompactValue(sourceProbeContentTypeEl, "-");
    setCompactValue(sourceProbeBodyBytesEl, "-");
    setCompactValue(sourceProbeHeadersCountEl, "-");
    setCompactValue(sourceProbeDurationEl, "-");

    setCompactValue(sourceProbeDetectedFormatEl, "-");
    setCompactValue(sourceProbeFormatDetailEl, "-");

    setCompactValue(sourceProbeJSONValidEl, "-");
    setCompactValue(sourceProbeJSONTypeEl, "-");
    setCompactValue(sourceProbeListCountEl, "-");
    setCompactValue(sourceProbeEntryCountEl, "-");

    setCompactValue(sourceProbeTextLineCountEl, "-");
    setCompactValue(sourceProbeTextValidLineCountEl, "-");
    setCompactValue(sourceProbeTextInvalidLineCountEl, "-");

    setPreBlock(sourceProbeErrorEl, "尚未执行测试");
    setPreBlock(sourceProbeListNamesEl, "尚未执行测试");
    setPreBlock(sourceProbeResponseHeadersEl, "尚未执行测试");
    setPreBlock(sourceProbeRawPreviewEl, "尚未执行测试");
}

function renderSourceProbeResult(result) {
    lastProbeResult = result || null;

    setCompactValue(sourceProbeStatusEl, result?.ok ? "成功" : "失败");
    setCompactValue(sourceProbeNameEl, result?.name || "-");
    setCompactValue(sourceProbeLocationEl, result?.location || "-");
    setCompactValue(sourceProbeHTTPStatusEl, result?.status_text || "-");
    setCompactValue(sourceProbeContentTypeEl, result?.content_type || "-");
    setCompactValue(sourceProbeBodyBytesEl, result?.body_bytes ?? 0);
    setCompactValue(sourceProbeHeadersCountEl, result?.headers_count ?? 0);
    setCompactValue(sourceProbeDurationEl, `${result?.duration_ms ?? 0} ms`);

    setCompactValue(sourceProbeDetectedFormatEl, result?.detected_format || "-");
    setCompactValue(sourceProbeFormatDetailEl, result?.format_detail || "-");

    setCompactValue(sourceProbeJSONValidEl, String(Boolean(result?.json_valid)));
    setCompactValue(sourceProbeJSONTypeEl, result?.json_type || "-");

    setCompactValue(sourceProbeListCountEl, result?.detected_list_count ?? 0);
    setCompactValue(sourceProbeEntryCountEl, result?.detected_entry_count ?? 0);

    setCompactValue(sourceProbeTextLineCountEl, result?.text_line_count ?? 0);
    setCompactValue(sourceProbeTextValidLineCountEl, result?.text_valid_line_count ?? 0);
    setCompactValue(sourceProbeTextInvalidLineCountEl, result?.text_invalid_line_count ?? 0);

    const warnings = Array.isArray(result?.warnings) ? result.warnings : [];
    const errorText = result?.error
        ? [result.error, ...warnings].join("\n\n")
        : warnings.join("\n") || "无";
    setPreBlock(sourceProbeErrorEl, errorText, "无");

    const listNames = Array.isArray(result?.list_names) ? result.list_names : [];
    setPreBlock(sourceProbeListNamesEl, listNames.length ? listNames.join("\n") : "-", "-");

    setPreBlock(
        sourceProbeResponseHeadersEl,
        safeJSONStringify(result?.response_headers || {}),
        "{}"
    );

    setPreBlock(sourceProbeRawPreviewEl, result?.raw_preview || "", "空响应");
}

async function testSourceWithPayload(kind, payload) {
    clearMessage();

    if (!payload.name) {
        showMessage("Source Name 不能为空。", true);
        return;
    }

    if (payload.type === "file" && !payload.path) {
        showMessage("type=file 时 Path 不能为空。", true);
        return;
    }

    if (payload.type === "url" && !payload.url) {
        showMessage("type=url 时 URL 不能为空。", true);
        return;
    }

    setSourceProbeLoading(kind, payload);

    try {
        const result = await apiFetch("/api/v1/sources/test", {
            method: "POST",
            body: JSON.stringify(payload),
        });

        renderSourceProbeResult(result);

        const fmt = result?.detected_format || "unknown";
        showMessage(`Source 测试完成：${payload.name}（检测格式：${fmt}）`);
    } catch (error) {
        renderSourceProbeResult({
            ok: false,
            name: payload.name,
            location: payload.type === "file" ? payload.path : payload.url,
            error: error.message,
        });
        showMessage(`Source 测试失败：${error.message}`, true);
    }
}

async function saveLastProbedSource() {
    clearMessage();

    if (!lastProbePayload || !lastProbeKind) {
        showMessage("当前没有可保存的测试参数，请先执行一次 Source 测试。", true);
        return;
    }

    const payload = { ...lastProbePayload };

    const path =
        lastProbeKind === "desired"
            ? "/api/v1/sources/desired"
            : "/api/v1/sources/current";

    try {
        await apiFetch(path, {
            method: "POST",
            body: JSON.stringify(payload),
        });

        if (lastProbeKind === "desired") {
            await loadDesiredSources();
        } else {
            await loadCurrentSources();
        }
        await loadConfig();

        showMessage(`已保存本次测试参数到 ${lastProbeKind} sources：${payload.name}`);
    } catch (error) {
        showMessage(`保存测试参数失败：${error.message}`, true);
    }
}

// Render
function clearRenderResultView() {
    currentRenderResult = null;
    setCompactValue(renderResultModeEl, "-");
    setCompactValue(renderResultListCountEl, "-");
    setCompactValue(renderResultEntryCountEl, "-");
    setCompactValue(renderResultOutputPathEl, "-");
    setPreBlock(renderScriptViewerEl, "尚未执行渲染");
}

function resetRenderForm() {
    if (renderForm) renderForm.reset();
    renderModeSelect.value = "";
    renderOutputPathInput.value = currentConfig?.output?.path || "";
}

async function executeRender(event) {
    event.preventDefault();
    clearMessage();

    const payload = {
        mode: renderModeSelect.value || "",
        output_path: renderOutputPathInput.value.trim(),
    };

    try {
        const result = await apiFetch("/api/v1/render", {
            method: "POST",
            body: JSON.stringify(payload),
        });

        currentRenderResult = result;
        setCompactValue(renderResultModeEl, result?.mode || "-");
        setCompactValue(renderResultListCountEl, result?.list_count ?? 0);
        setCompactValue(renderResultEntryCountEl, result?.entry_count ?? 0);
        setCompactValue(renderResultOutputPathEl, result?.output_path || "-");
        setPreBlock(renderScriptViewerEl, result?.script || "", "空脚本");

        showMessage("渲染执行完成。");
    } catch (error) {
        showMessage(`执行渲染失败：${error.message}`, true);
    }
}

async function copyRenderScript() {
    clearMessage();

    const script = currentRenderResult?.script || renderScriptViewerEl?.textContent || "";
    if (!script || script === "尚未执行渲染") {
        showMessage("当前没有可复制的脚本。", true);
        return;
    }

    try {
        await navigator.clipboard.writeText(script);
        showMessage("脚本已复制到剪贴板。");
    } catch (error) {
        showMessage(`复制失败：${error.message}`, true);
    }
}

// Bindings
navButtons.forEach((button) => {
    button.addEventListener("click", () => {
        const sectionId = button.dataset.section;
        if (sectionId) {
            setActiveSection(sectionId);
        }
    });
});

if (btnCheckHealth) {
    btnCheckHealth.onclick = checkHealth;
}

if (btnLoadConfig) {
    btnLoadConfig.onclick = async () => {
        clearMessage();
        await loadConfig();
        showMessage("配置已重新加载。");
    };
}

if (btnLoadConfigInline) {
    btnLoadConfigInline.onclick = async () => {
        clearMessage();
        await loadConfig();
        showMessage("配置已重新加载。");
    };
}

if (btnRefreshLists) {
    btnRefreshLists.onclick = async () => {
        clearMessage();
        await loadLists();
        showMessage("列表已刷新。");
    };
}

if (btnResetListForm) {
    btnResetListForm.onclick = () => {
        resetListForm();
        clearMessage();
    };
}

if (listForm) {
    listForm.addEventListener("submit", submitListForm);
}

if (descriptionForm) {
    descriptionForm.addEventListener("submit", submitDescriptionForm);
}

if (btnRefreshRules) {
    btnRefreshRules.onclick = async () => {
        clearMessage();
        await loadRules();
        showMessage("规则已刷新。");
    };
}

if (btnResetRuleForm) {
    btnResetRuleForm.onclick = () => {
        resetRuleForm();
        clearMessage();
    };
}

if (ruleForm) {
    ruleForm.addEventListener("submit", submitRuleForm);
}

if (btnRefreshDesiredSources) {
    btnRefreshDesiredSources.onclick = async () => {
        clearMessage();
        await loadDesiredSources();
        showMessage("Desired Sources 已刷新。");
    };
}

if (btnRefreshCurrentSources) {
    btnRefreshCurrentSources.onclick = async () => {
        clearMessage();
        await loadCurrentSources();
        showMessage("Current Sources 已刷新。");
    };
}

if (btnResetDesiredSourceForm) {
    btnResetDesiredSourceForm.onclick = () => {
        resetDesiredSourceForm();
        clearMessage();
    };
}

if (btnResetCurrentSourceForm) {
    btnResetCurrentSourceForm.onclick = () => {
        resetCurrentSourceForm();
        clearMessage();
    };
}

if (desiredSourceTypeSelect) {
    desiredSourceTypeSelect.addEventListener("change", () => syncSourceTypeUI("desired"));
}

if (currentSourceTypeSelect) {
    currentSourceTypeSelect.addEventListener("change", () => syncSourceTypeUI("current"));
}

if (desiredSourceForm) {
    desiredSourceForm.addEventListener("submit", submitDesiredSourceForm);
}

if (currentSourceForm) {
    currentSourceForm.addEventListener("submit", submitCurrentSourceForm);
}

if (btnTestDesiredSource) {
    btnTestDesiredSource.onclick = async () => {
        try {
            const payload = buildSourcePayloadFromDesiredForm();
            await testSourceWithPayload("desired", payload);
        } catch (error) {
            showMessage(`读取 Desired Source 表单失败：${error.message}`, true);
        }
    };
}

if (btnTestCurrentSource) {
    btnTestCurrentSource.onclick = async () => {
        try {
            const payload = buildSourcePayloadFromCurrentForm();
            await testSourceWithPayload("current", payload);
        } catch (error) {
            showMessage(`读取 Current Source 表单失败：${error.message}`, true);
        }
    };
}

if (btnSaveProbedSource) {
    btnSaveProbedSource.onclick = async () => {
        await saveLastProbedSource();
    };
}

if (btnClearSourceProbeResult) {
    btnClearSourceProbeResult.onclick = () => {
        clearMessage();
        clearSourceProbeResultView();
    };
}

if (btnResetRenderForm) {
    btnResetRenderForm.onclick = () => {
        resetRenderForm();
        clearMessage();
    };
}

if (btnClearRenderResult) {
    btnClearRenderResult.onclick = () => {
        clearMessage();
        clearRenderResultView();
    };
}

if (btnCopyRenderScript) {
    btnCopyRenderScript.onclick = copyRenderScript;
}

if (renderForm) {
    renderForm.addEventListener("submit", executeRender);
}


if (btnOpenChangePassword) {
    btnOpenChangePassword.onclick = () => {
        window.location.href = "/login.html?change_password=1";
    };
}

if (btnLogout) {
    btnLogout.onclick = async () => {
        try {
            await apiFetch("/api/v1/auth/logout", {
                method: "POST",
                body: JSON.stringify({}),
            });
        } catch (_) {
            // 即便接口失败，也尝试跳回登录页。
        }
        window.location.href = "/login.html";
    };
}

function resolveInitialSection() {
    const requestedSection = String(pageQuery.get("section") || "").trim();
    if (requestedSection && sections.some((section) => section.id === requestedSection)) {
        return requestedSection;
    }
    return "dashboard-section";
}

function getEntryAction() {
    return String(pageQuery.get("entry_action") || "").trim().toLowerCase();
}

function shouldAutoLoadDefaultConfig() {
    return getEntryAction() === "load_default_config";
}

function shouldAutoBootstrapOverview() {
    return getEntryAction() === "bootstrap_overview";
}

function clearEntryQueryFlags() {
    if (!window.history || typeof window.history.replaceState !== "function") {
        return;
    }

    const cleanURL = `${window.location.pathname}${window.location.hash || ""}`;
    window.history.replaceState({}, document.title, cleanURL);
}

// Init
async function initializePage() {
    await ensureAuthenticatedPage();

    clearMessage();
    clearSourceProbeResultView();
    clearRenderResultView();

    resetListForm();
    resetRuleForm();
    resetDesiredSourceForm();
    resetCurrentSourceForm();
    resetRenderForm();

    const initialSection = resolveInitialSection();
    setActiveSection(initialSection);

    await Promise.all([
        loadConfig(),
        loadLists(),
        loadRules(),
        loadDesiredSources(),
        loadCurrentSources(),
    ]);

    if (shouldAutoLoadDefaultConfig()) {
        setActiveSection("config-section");
        await loadConfig();
        showMessage("登录成功，已在后台自动加载默认配置。", false);
        clearEntryQueryFlags();
        return;
    }

    if (shouldAutoBootstrapOverview()) {
        try {
            await fetchHealthStatus();
            showMessage("登录成功，已在总览页自动检查服务状态并加载当前配置。", false);
        } catch (error) {
            setCompactValue(healthStatusEl, "失败");
            showMessage(`已加载当前配置，但服务状态检查失败：${error.message}`, true);
        }
        clearEntryQueryFlags();
    }
}

initializePage().catch((error) => {
    showMessage(`初始化页面失败：${error.message}`, true);
});