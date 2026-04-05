const navButtons = Array.from(document.querySelectorAll(".nav-item"));
const sections = Array.from(document.querySelectorAll(".page-section"));

const messageBox = document.getElementById("message-box");

const btnCheckHealth = document.getElementById("btn-check-health");
const btnLoadConfig = document.getElementById("btn-load-config");
const btnLoadConfigInline = document.getElementById("btn-load-config-inline");

const btnRefreshLists = document.getElementById("btn-refresh-lists");
const btnResetListForm = document.getElementById("btn-reset-list-form");

const healthStatusEl = document.getElementById("dashboard-health-status");
const listCountEl = document.getElementById("dashboard-list-count");
const ruleCountEl = document.getElementById("dashboard-rule-count");
const renderModeEl = document.getElementById("dashboard-render-mode");
const summaryListEl = document.getElementById("dashboard-summary-list");
const configViewerEl = document.getElementById("config-json-viewer");

const listsTableBody = document.getElementById("lists-table-body");

const listForm = document.getElementById("list-form");
const listNameInput = document.getElementById("list-name");
const listFamilySelect = document.getElementById("list-family");
const listEnabledInput = document.getElementById("list-enabled");
const listDescriptionInput = document.getElementById("list-description");

const descriptionForm = document.getElementById("description-form");
const descTargetNameInput = document.getElementById("desc-target-name");
const descTextInput = document.getElementById("desc-text");

let currentConfig = null;
let currentLists = [];
let editingListName = "";

// ===================== 页面导航 =====================

function setActiveSection(sectionId) {
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

// ===================== 消息提示 =====================

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

// ===================== API 请求层 =====================

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

// ===================== 基础功能 =====================

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
    clearMessage();

    try {
        const data = await apiFetch("/api/v1/config");
        currentConfig = data;

        renderConfigToDashboard(data);
        renderConfigToViewer(data);

        showMessage("配置加载成功。");
    } catch (error) {
        showMessage(`配置加载失败：${error.message}`, true);
    }
}

function renderConfigToDashboard(config) {
    listCountEl.textContent = String(config.lists?.length ?? 0);
    ruleCountEl.textContent = String(config.manual_rules?.length ?? 0);
    renderModeEl.textContent = String(config.output?.mode ?? "-");

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

// ===================== Address Lists 页面 =====================

async function loadLists() {
    clearMessage();

    try {
        const lists = await apiFetch("/api/v1/lists");
        currentLists = Array.isArray(lists) ? lists : [];
        renderListsTable(currentLists);

        // 如果配置也已经有了，顺手同步总览统计。
        if (currentConfig) {
            currentConfig.lists = currentLists;
            renderConfigToDashboard(currentConfig);
            renderConfigToViewer(currentConfig);
        }

        showMessage("Address Lists 已刷新。");
    } catch (error) {
        showMessage(`加载 Address Lists 失败：${error.message}`, true);
    }
}

function renderListsTable(lists) {
    listsTableBody.innerHTML = "";

    if (!lists.length) {
        const tr = document.createElement("tr");
        tr.innerHTML = `<td colspan="5">当前没有任何 list</td>`;
        listsTableBody.appendChild(tr);
        return;
    }

    lists.forEach((item) => {
        const tr = document.createElement("tr");

        const enabledText = item.enabled ? "true" : "false";
        const descriptionText = item.description || "";

        tr.innerHTML = `
      <td>${escapeHTML(item.name)}</td>
      <td>${escapeHTML(item.family)}</td>
      <td>${enabledText}</td>
      <td>${escapeHTML(descriptionText)}</td>
      <td>
        <div class="inline-actions">
          <button class="inline-link-btn" data-action="edit" data-name="${encodeURIComponent(item.name)}">编辑</button>
          <button class="inline-link-btn" data-action="desc" data-name="${encodeURIComponent(item.name)}">改说明</button>
          <button class="inline-link-btn danger" data-action="delete" data-name="${encodeURIComponent(item.name)}">删除</button>
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
        showMessage(`已加载 ${name} 到编辑表单。`);
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
        showMessage("Name 不能为空。", true);
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
        } else if (editingListName && editingListName !== name) {
            // 当前后端 PUT 的路径名优先，所以如果你改了 name，
            // 最简单稳妥的做法就是当作新增一个新 list。
            await apiFetch("/api/v1/lists", {
                method: "POST",
                body: JSON.stringify(payload)
            });
            showMessage(`已新增 list：${name}`);
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

function bindListTableActions() {
    listsTableBody.addEventListener("click", async (event) => {
        const target = event.target;
        if (!(target instanceof HTMLElement)) return;

        const action = target.dataset.action;
        const encodedName = target.dataset.name;
        if (!action || !encodedName) return;

        const name = decodeURIComponent(encodedName);

        if (action === "edit") {
            await editList(name);
            return;
        }

        if (action === "desc") {
            await editList(name);
            descTargetNameInput.value = name;
            setActiveSection("lists-section");
            return;
        }

        if (action === "delete") {
            await deleteList(name);
        }
    });
}

// ===================== 工具函数 =====================

function escapeHTML(value) {
    return String(value)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#39;");
}

// ===================== 事件绑定 =====================

btnCheckHealth.addEventListener("click", () => {
    void checkHealth();
});

btnLoadConfig.addEventListener("click", () => {
    void loadConfig();
});

btnLoadConfigInline.addEventListener("click", () => {
    void loadConfig();
    setActiveSection("config-section");
});

btnRefreshLists.addEventListener("click", () => {
    void loadLists();
});

btnResetListForm.addEventListener("click", () => {
    resetListForm();
    showMessage("List 表单已重置。");
});

listForm.addEventListener("submit", (event) => {
    void submitListForm(event);
});

descriptionForm.addEventListener("submit", (event) => {
    void submitDescriptionForm(event);
});

bindListTableActions();

// ===================== 页面初始化 =====================

window.addEventListener("DOMContentLoaded", async () => {
    setActiveSection("dashboard-section");

    await checkHealth();
    await loadConfig();
    await loadLists();
});