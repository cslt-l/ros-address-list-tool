/*
  第一版原生 JS 页面脚本设计目标：

  1. 不依赖任何前端框架
  2. 代码尽量直白易懂
  3. 先完成基础导航切换
  4. 先接通两个基础接口：
     - /healthz
     - /api/v1/config

  当前阶段暂不接 lists / rules / render 的完整交互，
  但页面结构已经预留。
*/

// ===================== 基础元素获取 =====================
const navButtons = Array.from(document.querySelectorAll(".nav-item"));
const sections = Array.from(document.querySelectorAll(".page-section"));

const messageBox = document.getElementById("message-box");

const btnCheckHealth = document.getElementById("btn-check-health");
const btnLoadConfig = document.getElementById("btn-load-config");
const btnLoadConfigInline = document.getElementById("btn-load-config-inline");

const healthStatusEl = document.getElementById("dashboard-health-status");
const listCountEl = document.getElementById("dashboard-list-count");
const ruleCountEl = document.getElementById("dashboard-rule-count");
const renderModeEl = document.getElementById("dashboard-render-mode");
const summaryListEl = document.getElementById("dashboard-summary-list");
const configViewerEl = document.getElementById("config-json-viewer");

// 当前页面缓存的配置对象。
// 后续 lists / rules / render 接上后，也可以继续复用。
let currentConfig = null;

// ===================== 页面导航 =====================

/*
  setActiveSection 用于切换当前显示的 section。

  为什么不用多页跳转？
  因为当前阶段我们要尽量轻量：
  - 不引入前端路由
  - 不引入构建工具
  - 直接用原生 JS 完成单页区块切换
*/
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

/*
  apiFetch 是当前页面统一的请求函数。

  这里故意使用“同源相对路径”：
  - /healthz
  - /api/v1/config

  原因是：
  下一步我们会让 Go 后端直接托管 web/dist，
  那时页面和 API 天然同源，不需要前端硬编码 127.0.0.1:9000。
*/
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

// ===================== 基础功能：健康检查 =====================

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

// ===================== 基础功能：加载完整配置 =====================

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

/*
  renderConfigToDashboard 用于把完整配置对象的关键摘要渲染到总览区。
*/
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

/*
  renderConfigToViewer 用于把完整配置对象以 JSON 字符串形式展示到配置查看区。
*/
function renderConfigToViewer(config) {
    configViewerEl.textContent = JSON.stringify(config, null, 2);
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

// ===================== 页面初始化 =====================

/*
  页面启动时，先自动检查一次服务状态，再自动尝试加载配置。
  这样打开页面后，用户能第一时间看到：
  - 服务通不通
  - 配置能不能拿到
*/
window.addEventListener("DOMContentLoaded", async () => {
    setActiveSection("dashboard-section");

    await checkHealth();
    await loadConfig();
});