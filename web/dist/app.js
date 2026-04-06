const navButtons = Array.from(document.querySelectorAll(".nav-item"));
const sections = Array.from(document.querySelectorAll(".page-section"));

const messageBox = document.getElementById("message-box");

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

const renderForm = document.getElementById("render-form");
const renderModeSelect = document.getElementById("render-mode");
const renderOutputPathInput = document.getElementById("render-output-path");
const renderResultModeEl = document.getElementById("render-result-mode");
const renderResultListCountEl = document.getElementById("render-result-list-count");
const renderResultEntryCountEl = document.getElementById("render-result-entry-count");
const renderResultOutputPathEl = document.getElementById("render-result-output-path");
const renderScriptViewerEl = document.getElementById("render-script-viewer");

let currentConfig = null;
let currentLists = [];
let currentRules = [];
let currentDesiredSources = [];
let currentCurrentSources = [];

let editingListName = "";
let editingRuleId = "";
let editingDesiredSourceName = "";
let editingCurrentSourceName = "";

let currentRenderResult = null;
let lastProbeKind = "";
let lastProbePayload = null;
let lastProbeSourceName = "";

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

navButtons.forEach((button) => {
    button.addEventListener("click", () => {
        const sectionId = button.dataset.section;
        if (sectionId) {
            setActiveSection(sectionId);
        }
    });
});

function showMessage(message, isError = false) {
    messageBox.textContent = message;
    messageBox.classList.remove("hidden");
    messageBox.classList.toggle("error", isError);
}

function clearMessage() {
    messageBox.textContent = "";
    messageBox.classList.add("hidden");
    messageBox.classList.remove("error");
}

async function apiFetch(path, options = {}) {
    const response = await fetch(path, {
        headers: {
            "Content-Type": "application/json; charset=utf-8",
            ...(options.headers || {})
        },
        ...options
    });

    const text = await response.text();
    let data = null;

    try {
        data = text ? JSON.parse(text) : null;
    } catch (err) {
        throw new Error(`响应不是合法 JSON：${text}`);
    }

    if (!response.ok) {
        const message =
            data && typeof data.error === "string"
                ? data.error
                : `HTTP ${response.status}`;
        throw new Error(message);
    }

    return data;
}

async function checkHealth() {
    clearMessage();

    try {
        healthStatusEl.textContent = "检查中...";
        const data = await apiFetch("/healthz");
        healthStatusEl.textContent = data.status || "unknown";
        showMessage("服务状态检查成功。");
    } catch (error) {
        healthStatusEl.textContent = "失败";
        showMessage(`服务状态检查失败：${error.message}`, true);
    }
}

async function loadConfig() {
    try {
        const data = await apiFetch("/api/v1/config");
        currentConfig = data;
        renderConfigToDashboard(data);
        renderConfigToViewer(data);
    } catch (error) {
        showMessage(`配置加载失败：${error.message}`, true);
    }
}

function renderConfigToDashboard(config) {
    listCountEl.textContent = String(config.lists?.length ?? 0);
    ruleCountEl.textContent = String(config.manual_rules?.length ?? 0);
    renderModeEl.textContent = String(config.output?.mode ?? "-");
    desiredSourceCountEl.textContent = String(config.desired_sources?.length ?? 0);
    currentSourceCountEl.textContent = String(config.current_state_sources?.length ?? 0);

    const summaryItems = [
        `自动创建未知 list：${String(config.auto_create_lists)}`,
        `日志文件：${config.log_file ?? "-"}`,
        `输出路径：${config.output?.path ?? "-"}`,
        `managed comment：${config.output?.managed_comment ?? "-"}`,
        `监听地址：${config.server?.listen ?? "-"}`,
        `desired_sources 数量：${config.desired_sources?.length ?? 0}`,
        `current_state_sources 数量：${config.current_state_sources?.length ?? 0}`
    ];

    summaryListEl.innerHTML = "";
    summaryItems.forEach((item) => {
        const li = document.createElement("li");
        li.textContent = item;
        summaryListEl.appendChild(li);
    });
}

function renderConfigToViewer(config) {
    configViewerEl.textContent = JSON.stringify(config, null, 2);
}

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
    listsTableBody.innerHTML = "";

    if (!lists.length) {
        const tr = document.createElement("tr");
        tr.innerHTML = `<td colspan="5" class="empty-cell">当前没有任何 list</td>`;
        listsTableBody.appendChild(tr);
        return;
    }

    lists.forEach((item) => {
        const tr = document.createElement("tr");
        const enabledText = item.enabled ? "true" : "false";

        tr.innerHTML = `
      <td>${escapeHTML(item.name)}</td>
      <td>${escapeHTML(item.family)}</td>
      <td>${enabledText}</td>
      <td>${escapeHTML(item.description || "")}</td>
      <td>
        <div class="inline-actions">
          <button class="inline-link-btn" data-action="edit-list" data-name="${encodeURIComponent(item.name)}">编辑</button>
          <button class="inline-link-btn" data-action="desc-list" data-name="${encodeURIComponent(item.name)}">改说明</button>
          <button class="inline-link-btn danger" data-action="delete-list" data-name="${encodeURIComponent(item.name)}">删除</button>
        </div>
      </td>
    `;

        listsTableBody.appendChild(tr);
    });
}

async function getListByName(name) {
    return apiFetch(`/api/v1/lists/${encodeURIComponent(name)}`);
}

async function editList(name) {
    clearMessage();

    try {
        const item = await getListByName(name);

        editingListName = item.name;
        listNameInput.value = item.name || "";
        listFamilySelect.value = item.family || "ipv4";
        listEnabledInput.checked = Boolean(item.enabled);
        listDescriptionInput.value = item.description || "";

        descTargetNameInput.value = item.name || "";
        descTextInput.value = item.description || "";

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
            method: "DELETE"
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
    listForm.reset();
    listFamilySelect.value = "ipv4";
    listEnabledInput.checked = true;
}

async function submitListForm(event) {
    event.preventDefault();
    clearMessage();

    const name = listNameInput.value.trim();
    const family = listFamilySelect.value;
    const enabled = listEnabledInput.checked;
    const description = listDescriptionInput.value;

    if (!name) {
        showMessage("List Name 不能为空。", true);
        return;
    }

    const payload = {
        name,
        family,
        enabled,
        description
    };

    try {
        if (editingListName && editingListName === name) {
            await apiFetch(`/api/v1/lists/${encodeURIComponent(name)}`, {
                method: "PUT",
                body: JSON.stringify(payload)
            });
            showMessage(`已更新 list：${name}`);
        } else {
            await apiFetch("/api/v1/lists", {
                method: "POST",
                body: JSON.stringify(payload)
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
            body: JSON.stringify({ description })
        });

        await loadLists();
        await loadConfig();

        showMessage(`已更新 ${name} 的 description。`);
    } catch (error) {
        showMessage(`更新 description 失败：${error.message}`, true);
    }
}

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
    rulesTableBody.innerHTML = "";

    if (!rules.length) {
        const tr = document.createElement("tr");
        tr.innerHTML = `<td colspan="8" class="empty-cell">当前没有任何 rule</td>`;
        rulesTableBody.appendChild(tr);
        return;
    }

    rules.forEach((item) => {
        const tr = document.createElement("tr");
        const enabledText = item.enabled ? "true" : "false";
        const entriesText = Array.isArray(item.entries) ? item.entries.join(", ") : "";

        tr.innerHTML = `
      <td>${escapeHTML(item.id)}</td>
      <td>${escapeHTML(item.list_name)}</td>
      <td>${escapeHTML(item.action)}</td>
      <td>${escapeHTML(String(item.priority))}</td>
      <td>${enabledText}</td>
      <td>${escapeHTML(item.description || "")}</td>
      <td>${escapeHTML(entriesText)}</td>
      <td>
        <div class="inline-actions">
          <button class="inline-link-btn" data-action="edit-rule" data-id="${encodeURIComponent(item.id)}">编辑</button>
          <button class="inline-link-btn danger" data-action="delete-rule" data-id="${encodeURIComponent(item.id)}">删除</button>
        </div>
      </td>
    `;

        rulesTableBody.appendChild(tr);
    });
}

function resetRuleForm() {
    editingRuleId = "";
    ruleForm.reset();
    ruleActionSelect.value = "add";
    rulePriorityInput.value = "1000";
    ruleEnabledInput.checked = true;
}

function fillRuleForm(rule) {
    editingRuleId = rule.id || "";
    ruleIdInput.value = rule.id || "";
    ruleListNameInput.value = rule.list_name || "";
    ruleActionSelect.value = rule.action || "add";
    rulePriorityInput.value = String(rule.priority ?? 1000);
    ruleEnabledInput.checked = Boolean(rule.enabled);
    ruleDescriptionInput.value = rule.description || "";
    ruleEntriesInput.value = Array.isArray(rule.entries) ? rule.entries.join("\n") : "";
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
            method: "DELETE"
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
    const enabled = ruleEnabledInput.checked;
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
        entries
    };

    try {
        if (editingRuleId) {
            await apiFetch(`/api/v1/manual-rules/${encodeURIComponent(editingRuleId)}`, {
                method: "PUT",
                body: JSON.stringify(payload)
            });
            showMessage(`已更新 rule：${editingRuleId}`);
        } else {
            await apiFetch("/api/v1/manual-rules", {
                method: "POST",
                body: JSON.stringify(payload)
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

function resetDesiredSourceForm() {
    editingDesiredSourceName = "";
    desiredSourceForm.reset();
    desiredSourceHeadersInput.value = "";
    desiredSourceTypeSelect.value = "file";
    desiredSourcePriorityInput.value = "100";
    desiredSourceTimeoutInput.value = "15";
    desiredSourceEnabledInput.checked = true;
    syncSourceTypeUI("desired");
}

function resetCurrentSourceForm() {
    editingCurrentSourceName = "";
    currentSourceForm.reset();
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
    targetBody.innerHTML = "";

    if (!items.length) {
        const tr = document.createElement("tr");
        tr.innerHTML = `<td colspan="8" class="empty-cell">当前没有任何 source</td>`;
        targetBody.appendChild(tr);
        return;
    }

    items.forEach((item) => {
        const tr = document.createElement("tr");
        const location = item.type === "url" ? item.url || "" : item.path || "";
        const typeClass = item.type === "url" ? "source-type-badge url" : "source-type-badge";
        const headerCount = item.headers ? Object.keys(item.headers).length : 0;
        const headerText =
            headerCount > 0
                ? `<span class="header-count-badge">${headerCount} 个</span>`
                : "-";

        const toggleText = item.enabled ? "停用" : "启用";
        const enabledText = item.enabled ? "true" : "false";

        tr.innerHTML = `
      <td>${escapeHTML(item.name)}</td>
      <td><span class="${typeClass}">${escapeHTML(item.type)}</span></td>
      <td>${escapeHTML(location)}</td>
      <td>${headerText}</td>
      <td>${enabledText}</td>
      <td>${escapeHTML(String(item.priority ?? ""))}</td>
      <td>${escapeHTML(String(item.timeout_seconds ?? ""))}</td>
      <td>
        <div class="inline-actions">
          <button class="inline-link-btn" data-action="edit-source" data-kind="${kind}" data-name="${encodeURIComponent(item.name)}">编辑</button>
          <button class="inline-link-btn" data-action="copy-source" data-kind="${kind}" data-name="${encodeURIComponent(item.name)}">复制</button>
          <button class="inline-link-btn" data-action="test-source" data-kind="${kind}" data-name="${encodeURIComponent(item.name)}">测试</button>
          <button class="inline-link-btn" data-action="toggle-source-enabled" data-kind="${kind}" data-name="${encodeURIComponent(item.name)}">${toggleText}</button>
          <button class="inline-link-btn" data-action="decrease-source-priority" data-kind="${kind}" data-name="${encodeURIComponent(item.name)}">-10</button>
          <button class="inline-link-btn" data-action="increase-source-priority" data-kind="${kind}" data-name="${encodeURIComponent(item.name)}">+10</button>
          <button class="inline-link-btn danger" data-action="delete-source" data-kind="${kind}" data-name="${encodeURIComponent(item.name)}">删除</button>
        </div>
      </td>
    `;

        targetBody.appendChild(tr);
    });
}

function findDesiredSourceByName(name) {
    return currentDesiredSources.find((item) => item.name === name) || null;
}

function findCurrentSourceByName(name) {
    return currentCurrentSources.find((item) => item.name === name) || null;
}

function makeCopiedSourceName(kind, originalName) {
    const items = kind === "desired" ? currentDesiredSources : currentCurrentSources;
    const existingNames = new Set(items.map((item) => item.name));

    const baseName = originalName.endsWith("-copy")
        ? originalName
        : `${originalName}-copy`;

    if (!existingNames.has(baseName)) {
        return baseName;
    }

    let index = 2;
    while (existingNames.has(`${baseName}-${index}`)) {
        index += 1;
    }

    return `${baseName}-${index}`;
}

function copySavedSourceToForm(kind, name) {
    const item =
        kind === "desired" ? findDesiredSourceByName(name) : findCurrentSourceByName(name);

    if (!item) {
        showMessage(`未找到要复制的 source：${name}`, true);
        return;
    }

    const copiedName = makeCopiedSourceName(kind, item.name || "source");

    if (kind === "desired") {
        editingDesiredSourceName = "";

        desiredSourceNameInput.value = copiedName;
        desiredSourceTypeSelect.value = item.type || "file";
        desiredSourcePathInput.value = item.path || "";
        desiredSourceURLInput.value = item.url || "";
        desiredSourceHeadersInput.value = stringifyHeadersMap(item.headers);
        desiredSourcePriorityInput.value = String(item.priority ?? 100);
        desiredSourceTimeoutInput.value = String(item.timeout_seconds ?? 15);
        desiredSourceEnabledInput.checked = Boolean(item.enabled);

        syncSourceTypeUI("desired");
    } else {
        editingCurrentSourceName = "";

        currentSourceNameInput.value = copiedName;
        currentSourceTypeSelect.value = item.type || "file";
        currentSourcePathInput.value = item.path || "";
        currentSourceURLInput.value = item.url || "";
        currentSourceHeadersInput.value = stringifyHeadersMap(item.headers);
        currentSourcePriorityInput.value = String(item.priority ?? 100);
        currentSourceTimeoutInput.value = String(item.timeout_seconds ?? 15);
        currentSourceEnabledInput.checked = Boolean(item.enabled);

        syncSourceTypeUI("current");
    }

    setActiveSection("sources-section");
    showMessage(`已将 ${name} 复制到 ${kind} source 表单，新名称为：${copiedName}`);
}

function fillDesiredSourceForm(item) {
    editingDesiredSourceName = item.name || "";
    desiredSourceNameInput.value = item.name || "";
    desiredSourceTypeSelect.value = item.type || "file";
    desiredSourcePathInput.value = item.path || "";
    desiredSourceURLInput.value = item.url || "";
    desiredSourceHeadersInput.value = stringifyHeadersMap(item.headers);
    desiredSourcePriorityInput.value = String(item.priority ?? 100);
    desiredSourceTimeoutInput.value = String(item.timeout_seconds ?? 15);
    desiredSourceEnabledInput.checked = Boolean(item.enabled);
    syncSourceTypeUI("desired");
}

function fillCurrentSourceForm(item) {
    editingCurrentSourceName = item.name || "";
    currentSourceNameInput.value = item.name || "";
    currentSourceTypeSelect.value = item.type || "file";
    currentSourcePathInput.value = item.path || "";
    currentSourceURLInput.value = item.url || "";
    currentSourceHeadersInput.value = stringifyHeadersMap(item.headers);
    currentSourcePriorityInput.value = String(item.priority ?? 100);
    currentSourceTimeoutInput.value = String(item.timeout_seconds ?? 15);
    currentSourceEnabledInput.checked = Boolean(item.enabled);
    syncSourceTypeUI("current");
}

async function editSource(kind, name) {
    const item =
        kind === "desired" ? findDesiredSourceByName(name) : findCurrentSourceByName(name);

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
}

async function testSavedSource(kind, name) {
    const item =
        kind === "desired" ? findDesiredSourceByName(name) : findCurrentSourceByName(name);

    if (!item) {
        showMessage(`未找到已保存 source：${name}`, true);
        return;
    }

    clearMessage();

    try {
        await runSourceProbe(item, kind, name);
        showMessage(`已对已保存 ${kind} source 执行测试：${name}`);
    } catch (error) {
        showMessage(`测试已保存 source 失败：${error.message}`, true);
    }
}

async function toggleSavedSourceEnabled(kind, name) {
    const item =
        kind === "desired" ? findDesiredSourceByName(name) : findCurrentSourceByName(name);

    if (!item) {
        showMessage(`未找到要切换启停的 source：${name}`, true);
        return;
    }

    const nextEnabled = !Boolean(item.enabled);
    const actionText = nextEnabled ? "启用" : "停用";

    clearMessage();

    const payload = {
        name: item.name,
        type: item.type,
        path: item.path,
        url: item.url,
        headers: item.headers,
        timeout_seconds: item.timeout_seconds,
        enabled: nextEnabled,
        priority: item.priority
    };

    try {
        if (kind === "desired") {
            await apiFetch(`/api/v1/sources/desired/${encodeURIComponent(name)}`, {
                method: "PUT",
                body: JSON.stringify(payload)
            });
            await loadDesiredSources();
        } else {
            await apiFetch(`/api/v1/sources/current/${encodeURIComponent(name)}`, {
                method: "PUT",
                body: JSON.stringify(payload)
            });
            await loadCurrentSources();
        }

        await loadConfig();

        showMessage(`已${actionText} ${kind} source：${name}`);
    } catch (error) {
        showMessage(`切换 source 启停失败：${error.message}`, true);
    }
}

async function adjustSavedSourcePriority(kind, name, delta) {
    const item =
        kind === "desired" ? findDesiredSourceByName(name) : findCurrentSourceByName(name);

    if (!item) {
        showMessage(`未找到要调整优先级的 source：${name}`, true);
        return;
    }

    const currentPriority = Number(item.priority ?? 0);
    const nextPriority = Math.max(0, currentPriority + delta);

    if (nextPriority === currentPriority) {
        showMessage(`source：${name} 的优先级已是最小值 0`);
        return;
    }

    clearMessage();

    const payload = {
        name: item.name,
        type: item.type,
        path: item.path,
        url: item.url,
        headers: item.headers,
        timeout_seconds: item.timeout_seconds,
        enabled: Boolean(item.enabled),
        priority: nextPriority
    };

    try {
        if (kind === "desired") {
            await apiFetch(`/api/v1/sources/desired/${encodeURIComponent(name)}`, {
                method: "PUT",
                body: JSON.stringify(payload)
            });
            await loadDesiredSources();
        } else {
            await apiFetch(`/api/v1/sources/current/${encodeURIComponent(name)}`, {
                method: "PUT",
                body: JSON.stringify(payload)
            });
            await loadCurrentSources();
        }

        await loadConfig();

        const actionText = delta > 0 ? `+${delta}` : `${delta}`;
        showMessage(`已调整 ${kind} source：${name} 的优先级 ${actionText}，当前为 ${nextPriority}`);
    } catch (error) {
        showMessage(`调整 source 优先级失败：${error.message}`, true);
    }
}

async function deleteSource(kind, name) {
    const ok = window.confirm(`确定删除 ${kind} source "${name}" 吗？`);
    if (!ok) return;

    clearMessage();

    const path =
        kind === "desired"
            ? `/api/v1/sources/desired/${encodeURIComponent(name)}`
            : `/api/v1/sources/current/${encodeURIComponent(name)}`;

    try {
        await apiFetch(path, { method: "DELETE" });

        if (kind === "desired" && editingDesiredSourceName === name) {
            resetDesiredSourceForm();
        }
        if (kind === "current" && editingCurrentSourceName === name) {
            resetCurrentSourceForm();
        }

        await loadDesiredSources();
        await loadCurrentSources();
        await loadConfig();

        showMessage(`已删除 ${kind} source：${name}`);
    } catch (error) {
        showMessage(`删除 source 失败：${error.message}`, true);
    }
}

function syncSourceTypeUI(kind) {
    const isDesired = kind === "desired";

    const typeSelect = isDesired ? desiredSourceTypeSelect : currentSourceTypeSelect;
    const pathInput = isDesired ? desiredSourcePathInput : currentSourcePathInput;
    const urlInput = isDesired ? desiredSourceURLInput : currentSourceURLInput;
    const pathHelp = isDesired ? desiredSourcePathHelp : currentSourcePathHelp;
    const urlHelp = isDesired ? desiredSourceURLHelp : currentSourceURLHelp;

    const currentType = typeSelect.value;

    if (currentType === "file") {
        pathInput.disabled = false;
        urlInput.disabled = true;

        pathInput.placeholder = "./data/source.json";
        urlInput.placeholder = "当前 file 模式下无需填写 URL";

        pathHelp.textContent = "当前为 file 模式，请填写本地路径。";
        pathHelp.classList.remove("field-help-muted");

        urlHelp.textContent = "当前 file 模式下 URL 输入已禁用；Headers 通常也不需要。";
        urlHelp.classList.add("field-help-muted");
        return;
    }

    pathInput.disabled = true;
    urlInput.disabled = false;

    pathInput.placeholder = "当前 url 模式下无需填写 Path";
    urlInput.placeholder = "https://example.com/source.json";

    pathHelp.textContent = "当前 url 模式下 Path 输入已禁用。";
    pathHelp.classList.add("field-help-muted");

    urlHelp.textContent = "当前为 url 模式，请填写远程地址；如需鉴权可额外填写 Headers。";
    urlHelp.classList.remove("field-help-muted");
}

function buildSourcePayload(kind) {
    const isDesired = kind === "desired";

    const name = isDesired ? desiredSourceNameInput.value.trim() : currentSourceNameInput.value.trim();
    const type = isDesired ? desiredSourceTypeSelect.value : currentSourceTypeSelect.value;
    const path = isDesired ? desiredSourcePathInput.value.trim() : currentSourcePathInput.value.trim();
    const url = isDesired ? desiredSourceURLInput.value.trim() : currentSourceURLInput.value.trim();
    const headersText = isDesired
        ? desiredSourceHeadersInput.value
        : currentSourceHeadersInput.value;

    const priority = Number(
        isDesired ? desiredSourcePriorityInput.value || 0 : currentSourcePriorityInput.value || 0
    );
    const timeoutSeconds = Number(
        isDesired ? desiredSourceTimeoutInput.value || 0 : currentSourceTimeoutInput.value || 0
    );
    const enabled = isDesired ? desiredSourceEnabledInput.checked : currentSourceEnabledInput.checked;

    const headers = parseHeadersText(headersText);

    return {
        name,
        type,
        path: type === "file" ? path || undefined : undefined,
        url: type === "url" ? url || undefined : undefined,
        headers: Object.keys(headers).length > 0 ? headers : undefined,
        timeout_seconds: timeoutSeconds,
        enabled,
        priority
    };
}

async function submitDesiredSourceForm(event) {
    event.preventDefault();
    clearMessage();

    const payload = buildSourcePayload("desired");
    await saveSourcePayload("desired", payload, false);
}

async function submitCurrentSourceForm(event) {
    event.preventDefault();
    clearMessage();

    const payload = buildSourcePayload("current");
    await saveSourcePayload("current", payload, false);
}

async function saveSourcePayload(kind, payload, fromProbe) {
    if (!payload.name) {
        showMessage(`${kind === "desired" ? "Desired" : "Current"} Source Name 不能为空。`, true);
        return;
    }

    if (!Number.isFinite(payload.priority) || payload.priority < 0) {
        showMessage(`${kind === "desired" ? "Desired" : "Current"} Source Priority 必须是大于等于 0 的数字。`, true);
        return;
    }

    if (!Number.isFinite(payload.timeout_seconds) || payload.timeout_seconds <= 0) {
        showMessage(`${kind === "desired" ? "Desired" : "Current"} Source Timeout Seconds 必须大于 0。`, true);
        return;
    }

    if (payload.type === "file" && !payload.path) {
        showMessage("file 类型 source 必须填写 path。", true);
        return;
    }

    if (payload.type === "url" && !payload.url) {
        showMessage("url 类型 source 必须填写 url。", true);
        return;
    }

    try {
        if (kind === "desired") {
            if (editingDesiredSourceName && editingDesiredSourceName === payload.name) {
                await apiFetch(`/api/v1/sources/desired/${encodeURIComponent(payload.name)}`, {
                    method: "PUT",
                    body: JSON.stringify(payload)
                });
            } else {
                await apiFetch("/api/v1/sources/desired", {
                    method: "POST",
                    body: JSON.stringify(payload)
                });
            }

            await loadDesiredSources();
            editingDesiredSourceName = payload.name;
        } else {
            if (editingCurrentSourceName && editingCurrentSourceName === payload.name) {
                await apiFetch(`/api/v1/sources/current/${encodeURIComponent(payload.name)}`, {
                    method: "PUT",
                    body: JSON.stringify(payload)
                });
            } else {
                await apiFetch("/api/v1/sources/current", {
                    method: "POST",
                    body: JSON.stringify(payload)
                });
            }

            await loadCurrentSources();
            editingCurrentSourceName = payload.name;
        }

        await loadConfig();

        showMessage(
            fromProbe
                ? `已一键保存本次测试参数到 ${kind} source：${payload.name}`
                : `已保存 ${kind} source：${payload.name}`
        );
    } catch (error) {
        showMessage(`保存 source 失败：${error.message}`, true);
    }
}

function clearSourceProbeResult() {
    lastProbeKind = "";
    lastProbePayload = null;
    lastProbeSourceName = "";
    btnSaveProbedSource.disabled = true;

    sourceProbeStatusEl.textContent = "-";
    sourceProbeStatusEl.classList.remove("status-ok", "status-fail");

    sourceProbeNameEl.textContent = "-";
    sourceProbeLocationEl.textContent = "-";
    sourceProbeHTTPStatusEl.textContent = "-";
    sourceProbeContentTypeEl.textContent = "-";
    sourceProbeBodyBytesEl.textContent = "-";
    sourceProbeJSONValidEl.textContent = "-";
    sourceProbeJSONTypeEl.textContent = "-";
    sourceProbeListCountEl.textContent = "-";
    sourceProbeEntryCountEl.textContent = "-";
    sourceProbeHeadersCountEl.textContent = "-";
    sourceProbeDurationEl.textContent = "-";

    sourceProbeErrorEl.textContent = "尚未执行测试";
    sourceProbeErrorEl.classList.add("empty-viewer");

    sourceProbeListNamesEl.textContent = "尚未执行测试";
    sourceProbeListNamesEl.classList.add("empty-viewer");

    sourceProbeResponseHeadersEl.textContent = "尚未执行测试";
    sourceProbeResponseHeadersEl.classList.add("empty-viewer");

    sourceProbeRawPreviewEl.textContent = "尚未执行测试";
    sourceProbeRawPreviewEl.classList.add("empty-viewer");
}

function renderSourceProbeResult(result) {
    sourceProbeStatusEl.textContent = result.ok ? "成功" : "失败";
    sourceProbeStatusEl.classList.toggle("status-ok", Boolean(result.ok));
    sourceProbeStatusEl.classList.toggle("status-fail", !result.ok);

    sourceProbeNameEl.textContent = result.name || "-";
    sourceProbeLocationEl.textContent = result.location || "-";
    sourceProbeHTTPStatusEl.textContent = result.status_text || "-";
    sourceProbeContentTypeEl.textContent = result.content_type || "-";
    sourceProbeBodyBytesEl.textContent = String(result.body_bytes ?? 0);
    sourceProbeJSONValidEl.textContent = result.json_valid ? "true" : "false";
    sourceProbeJSONTypeEl.textContent = result.json_type || "-";
    sourceProbeListCountEl.textContent = String(result.detected_list_count ?? 0);
    sourceProbeEntryCountEl.textContent = String(result.detected_entry_count ?? 0);
    sourceProbeHeadersCountEl.textContent = String(result.headers_count ?? 0);
    sourceProbeDurationEl.textContent = `${result.duration_ms ?? 0} ms`;

    const warnings = Array.isArray(result.warnings) ? result.warnings : [];
    const errorParts = [];
    if (result.error) {
        errorParts.push(`错误：${result.error}`);
    }
    if (warnings.length > 0) {
        errorParts.push(`警告：\n- ${warnings.join("\n- ")}`);
    }

    sourceProbeErrorEl.textContent = errorParts.length > 0 ? errorParts.join("\n\n") : "无";
    sourceProbeErrorEl.classList.toggle("empty-viewer", errorParts.length === 0);

    const listNames = Array.isArray(result.list_names) ? result.list_names : [];
    sourceProbeListNamesEl.textContent = listNames.length > 0 ? listNames.join("\n") : "未检测到 list 名称";
    sourceProbeListNamesEl.classList.toggle("empty-viewer", listNames.length === 0);

    const responseHeaders = result.response_headers
        ? JSON.stringify(result.response_headers, null, 2)
        : "无响应头（可能是 file 来源或服务未返回）";
    sourceProbeResponseHeadersEl.textContent = responseHeaders;
    sourceProbeResponseHeadersEl.classList.toggle("empty-viewer", !result.response_headers);

    const rawPreview = result.raw_preview || "无可预览内容";
    sourceProbeRawPreviewEl.textContent = rawPreview;
    sourceProbeRawPreviewEl.classList.toggle("empty-viewer", !result.raw_preview);
}

async function runSourceProbe(payload, kind, sourceName = "") {
    const result = await apiFetch("/api/v1/sources/test", {
        method: "POST",
        body: JSON.stringify(payload)
    });

    renderSourceProbeResult(result);
    lastProbeKind = kind;
    lastProbePayload = payload;
    lastProbeSourceName = sourceName || payload.name || "";
    btnSaveProbedSource.disabled = false;

    if (result.ok) {
        showMessage("Source 测试成功。");
    } else {
        showMessage(`Source 测试失败：${result.error || "请查看测试结果区域"}`, true);
    }
}

async function testDesiredSource() {
    clearMessage();

    try {
        const payload = buildSourcePayload("desired");
        if (!payload.type) {
            showMessage("Desired Source Type 不能为空。", true);
            return;
        }

        if (payload.type === "file" && !payload.path) {
            showMessage("file 类型 source 必须填写 path。", true);
            return;
        }

        if (payload.type === "url" && !payload.url) {
            showMessage("url 类型 source 必须填写 url。", true);
            return;
        }

        await runSourceProbe(payload, "desired", payload.name || "");
    } catch (error) {
        showMessage(`测试 Desired Source 失败：${error.message}`, true);
    }
}

async function testCurrentSource() {
    clearMessage();

    try {
        const payload = buildSourcePayload("current");
        if (!payload.type) {
            showMessage("Current Source Type 不能为空。", true);
            return;
        }

        if (payload.type === "file" && !payload.path) {
            showMessage("file 类型 source 必须填写 path。", true);
            return;
        }

        if (payload.type === "url" && !payload.url) {
            showMessage("url 类型 source 必须填写 url。", true);
            return;
        }

        await runSourceProbe(payload, "current", payload.name || "");
    } catch (error) {
        showMessage(`测试 Current Source 失败：${error.message}`, true);
    }
}

async function saveProbedSource() {
    if (!lastProbeKind || !lastProbePayload) {
        showMessage("当前没有可保存的测试参数。", true);
        return;
    }

    await saveSourcePayload(lastProbeKind, lastProbePayload, true);
}

function bindListTableActions() {
    listsTableBody.addEventListener("click", async (event) => {
        const target = event.target;
        if (!(target instanceof HTMLElement)) return;

        const action = target.dataset.action;
        const encodedName = target.dataset.name;
        if (!action || !encodedName) return;

        const name = decodeURIComponent(encodedName);

        if (action === "edit-list") {
            await editList(name);
            return;
        }

        if (action === "desc-list") {
            await editList(name);
            descTargetNameInput.value = name;
            setActiveSection("lists-section");
            return;
        }

        if (action === "delete-list") {
            await deleteList(name);
        }
    });
}

function bindRuleTableActions() {
    rulesTableBody.addEventListener("click", async (event) => {
        const target = event.target;
        if (!(target instanceof HTMLElement)) return;

        const action = target.dataset.action;
        const encodedId = target.dataset.id;
        if (!action || !encodedId) return;

        const id = decodeURIComponent(encodedId);

        if (action === "edit-rule") {
            await editRule(id);
            return;
        }

        if (action === "delete-rule") {
            await deleteRule(id);
        }
    });
}

function bindSourceTableActions() {
    desiredSourcesTableBody.addEventListener("click", async (event) => {
        const target = event.target;
        if (!(target instanceof HTMLElement)) return;

        const action = target.dataset.action;
        const kind = target.dataset.kind;
        const encodedName = target.dataset.name;
        if (!action || !kind || !encodedName) return;

        const name = decodeURIComponent(encodedName);

        if (action === "edit-source") {
            await editSource(kind, name);
            return;
        }

        if (action === "copy-source") {
            copySavedSourceToForm(kind, name);
            return;
        }

        if (action === "test-source") {
            await testSavedSource(kind, name);
            return;
        }

        if (action === "toggle-source-enabled") {
            await toggleSavedSourceEnabled(kind, name);
            return;
        }

        if (action === "decrease-source-priority") {
            await adjustSavedSourcePriority(kind, name, -10);
            return;
        }

        if (action === "increase-source-priority") {
            await adjustSavedSourcePriority(kind, name, 10);
            return;
        }

        if (action === "delete-source") {
            await deleteSource(kind, name);
        }
    });

    currentSourcesTableBody.addEventListener("click", async (event) => {
        const target = event.target;
        if (!(target instanceof HTMLElement)) return;

        const action = target.dataset.action;
        const kind = target.dataset.kind;
        const encodedName = target.dataset.name;
        if (!action || !kind || !encodedName) return;

        const name = decodeURIComponent(encodedName);

        if (action === "edit-source") {
            await editSource(kind, name);
            return;
        }

        if (action === "copy-source") {
            copySavedSourceToForm(kind, name);
            return;
        }

        if (action === "test-source") {
            await testSavedSource(kind, name);
            return;
        }

        if (action === "toggle-source-enabled") {
            await toggleSavedSourceEnabled(kind, name);
            return;
        }

        if (action === "decrease-source-priority") {
            await adjustSavedSourcePriority(kind, name, -10);
            return;
        }

        if (action === "increase-source-priority") {
            await adjustSavedSourcePriority(kind, name, 10);
            return;
        }

        if (action === "delete-source") {
            await deleteSource(kind, name);
        }
    });
}

function resetRenderForm() {
    renderForm.reset();
    renderModeSelect.value = "";
    renderOutputPathInput.value = "";
}

function clearRenderResult() {
    currentRenderResult = null;
    renderResultModeEl.textContent = "-";
    renderResultListCountEl.textContent = "-";
    renderResultEntryCountEl.textContent = "-";
    renderResultOutputPathEl.textContent = "-";
    renderScriptViewerEl.textContent = "尚未执行渲染";
    renderScriptViewerEl.classList.add("empty-viewer");
}

function renderRenderResult(result) {
    currentRenderResult = result;
    renderResultModeEl.textContent = String(result.mode ?? "-");
    renderResultListCountEl.textContent = String(result.list_count ?? "-");
    renderResultEntryCountEl.textContent = String(result.entry_count ?? "-");
    renderResultOutputPathEl.textContent = String(result.output_path ?? "-");
    renderScriptViewerEl.textContent = result.script || "";
    renderScriptViewerEl.classList.remove("empty-viewer");
}

async function submitRenderForm(event) {
    event.preventDefault();
    clearMessage();

    const mode = renderModeSelect.value;
    const outputPath = renderOutputPathInput.value.trim();

    const payload = {};
    if (mode) {
        payload.mode = mode;
    }
    if (outputPath) {
        payload.output_path = outputPath;
    }

    try {
        const result = await apiFetch("/api/v1/render", {
            method: "POST",
            body: JSON.stringify(payload)
        });

        renderRenderResult(result);
        showMessage("渲染执行成功。");
    } catch (error) {
        showMessage(`渲染执行失败：${error.message}`, true);
    }
}

async function copyRenderScript() {
    const text = renderScriptViewerEl.textContent || "";
    if (!text || text === "尚未执行渲染") {
        showMessage("当前没有可复制的脚本。", true);
        return;
    }

    try {
        await navigator.clipboard.writeText(text);
        showMessage("脚本已复制到剪贴板。");
    } catch (error) {
        showMessage("复制失败，请手动复制。", true);
    }
}

function escapeHTML(value) {
    return String(value)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#39;");
}

btnCheckHealth.addEventListener("click", () => {
    void checkHealth();
});

btnLoadConfig.addEventListener("click", () => {
    void loadConfig().then(() => showMessage("配置加载成功。"));
});

btnLoadConfigInline.addEventListener("click", () => {
    void loadConfig().then(() => {
        setActiveSection("config-section");
        showMessage("配置加载成功。");
    });
});

btnRefreshLists.addEventListener("click", () => {
    clearMessage();
    void loadLists().then(() => showMessage("Address Lists 已刷新。"));
});

btnResetListForm.addEventListener("click", () => {
    resetListForm();
    showMessage("List 表单已重置。");
});

btnRefreshRules.addEventListener("click", () => {
    clearMessage();
    void loadRules().then(() => showMessage("Manual Rules 已刷新。"));
});

btnResetRuleForm.addEventListener("click", () => {
    resetRuleForm();
    showMessage("Rule 表单已重置。");
});

btnRefreshDesiredSources.addEventListener("click", () => {
    clearMessage();
    void loadDesiredSources().then(() => showMessage("Desired Sources 已刷新。"));
});

btnResetDesiredSourceForm.addEventListener("click", () => {
    resetDesiredSourceForm();
    showMessage("Desired Source 表单已重置。");
});

btnRefreshCurrentSources.addEventListener("click", () => {
    clearMessage();
    void loadCurrentSources().then(() => showMessage("Current Sources 已刷新。"));
});

btnResetCurrentSourceForm.addEventListener("click", () => {
    resetCurrentSourceForm();
    showMessage("Current Source 表单已重置。");
});

btnTestDesiredSource.addEventListener("click", () => {
    void testDesiredSource();
});

btnTestCurrentSource.addEventListener("click", () => {
    void testCurrentSource();
});

btnSaveProbedSource.addEventListener("click", () => {
    void saveProbedSource();
});

btnClearSourceProbeResult.addEventListener("click", () => {
    clearSourceProbeResult();
    showMessage("Source 测试结果已清空。");
});

desiredSourceTypeSelect.addEventListener("change", () => {
    syncSourceTypeUI("desired");
});

currentSourceTypeSelect.addEventListener("change", () => {
    syncSourceTypeUI("current");
});

btnResetRenderForm.addEventListener("click", () => {
    resetRenderForm();
    showMessage("渲染参数已重置。");
});

btnClearRenderResult.addEventListener("click", () => {
    clearRenderResult();
    showMessage("渲染结果已清空。");
});

btnCopyRenderScript.addEventListener("click", () => {
    void copyRenderScript();
});

listForm.addEventListener("submit", (event) => {
    void submitListForm(event);
});

descriptionForm.addEventListener("submit", (event) => {
    void submitDescriptionForm(event);
});

ruleForm.addEventListener("submit", (event) => {
    void submitRuleForm(event);
});

desiredSourceForm.addEventListener("submit", (event) => {
    void submitDesiredSourceForm(event);
});

currentSourceForm.addEventListener("submit", (event) => {
    void submitCurrentSourceForm(event);
});

renderForm.addEventListener("submit", (event) => {
    void submitRenderForm(event);
});

bindListTableActions();
bindRuleTableActions();
bindSourceTableActions();

window.addEventListener("DOMContentLoaded", async () => {
    setActiveSection("dashboard-section");
    resetListForm();
    resetRuleForm();
    resetDesiredSourceForm();
    resetCurrentSourceForm();
    resetRenderForm();
    clearRenderResult();
    clearSourceProbeResult();

    await checkHealth();
    await loadConfig();
    await loadLists();
    await loadRules();
    await loadDesiredSources();
    await loadCurrentSources();
});