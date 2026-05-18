// DOM 元素
const openApiUrlInput = document.getElementById('openApiUrl');
const parseBtn = document.getElementById('parseBtn');
const parseMessage = document.getElementById('parseMessage');
const selectionSection = document.getElementById('selectionSection');
const tagsContainer = document.getElementById('tagsContainer');
const selectAllBtn = document.getElementById('selectAllBtn');
const deselectAllBtn = document.getElementById('deselectAllBtn');
const downloadBtn = document.getElementById('downloadBtn');
const downloadMessage = document.getElementById('downloadMessage');
const themeToggle = document.getElementById('themeToggle');
const infoBar = document.getElementById('infoBar');

let allTags = [];
let currentVersion = '';

// 初始化主题
initializeTheme();

// 事件监听
parseBtn.addEventListener('click', handleParse);
selectAllBtn.addEventListener('click', selectAllTags);
deselectAllBtn.addEventListener('click', deselectAllTags);
downloadBtn.addEventListener('click', handleDownload);
themeToggle.addEventListener('click', toggleTheme);
openApiUrlInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') handleParse();
});

function initializeTheme() {
    const isDarkMode = localStorage.getItem('darkMode') === 'true';
    if (isDarkMode) {
        document.body.classList.add('dark-mode');
        themeToggle.textContent = '☀️';
    } else {
        document.body.classList.remove('dark-mode');
        themeToggle.textContent = '🌙';
    }
}

function toggleTheme() {
    const isDarkMode = document.body.classList.toggle('dark-mode');
    localStorage.setItem('darkMode', isDarkMode);
    themeToggle.textContent = isDarkMode ? '☀️' : '🌙';
}

function showMessage(element, message, type) {
    element.textContent = message;
    element.className = `message show ${type}`;

    if (type === 'success') {
        setTimeout(() => {
            element.classList.remove('show');
        }, 5000);
    }
}

function hideMessage(element) {
    element.classList.remove('show');
}

// 安全地设置信息栏内容（使用 DOM 方法避免 XSS）
function setInfoBarContent(version, tagCount) {
    infoBar.textContent = '';
    const text = document.createTextNode('已加载 ');
    infoBar.appendChild(text);

    const versionStrong = document.createElement('strong');
    versionStrong.textContent = `OpenAPI ${version}`;
    infoBar.appendChild(versionStrong);

    const text2 = document.createTextNode(' 文档，共 ');
    infoBar.appendChild(text2);

    const countStrong = document.createElement('strong');
    countStrong.textContent = String(tagCount);
    infoBar.appendChild(countStrong);

    const text3 = document.createTextNode(' 个模块');
    infoBar.appendChild(text3);
}

async function handleParse() {
    const url = openApiUrlInput.value.trim();

    if (!url) {
        showMessage(parseMessage, '请输入有效的 URL', 'error');
        return;
    }

    parseBtn.disabled = true;
    parseBtn.classList.add('loading');
    parseBtn.textContent = '加载中...';
    hideMessage(parseMessage);
    selectionSection.style.display = 'none';

    try {
        const response = await fetch('/api/parse', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ url }),
        });

        const data = await response.json();

        if (data.success) {
            allTags = data.tags || [];
            currentVersion = data.version || '';
            renderTags();
            selectionSection.style.display = 'block';
            setInfoBarContent(data.version, allTags.length);
            showMessage(parseMessage, 'OpenAPI 解析成功！', 'success');
        } else {
            showMessage(parseMessage, data.message || '解析 OpenAPI 失败', 'error');
        }
    } catch (error) {
        showMessage(parseMessage, `错误：${error.message}`, 'error');
    } finally {
        parseBtn.disabled = false;
        parseBtn.classList.remove('loading');
        parseBtn.textContent = '加载并解析';
    }
}

function renderTags() {
    tagsContainer.textContent = '';

    allTags.forEach((tag, index) => {
        const tagItem = document.createElement('label');
        tagItem.className = 'tag-item';

        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.value = tag;
        checkbox.addEventListener('change', updateDownloadButton);

        const span = document.createElement('span');
        span.textContent = tag;

        tagItem.appendChild(checkbox);
        tagItem.appendChild(span);
        tagsContainer.appendChild(tagItem);
    });
}

function selectAllTags() {
    const checkboxes = tagsContainer.querySelectorAll('input[type="checkbox"]');
    checkboxes.forEach(checkbox => {
        checkbox.checked = true;
    });
    updateDownloadButton();
}

function deselectAllTags() {
    const checkboxes = tagsContainer.querySelectorAll('input[type="checkbox"]');
    checkboxes.forEach(checkbox => {
        checkbox.checked = false;
    });
    updateDownloadButton();
}

function updateDownloadButton() {
    const checkboxes = tagsContainer.querySelectorAll('input[type="checkbox"]');
    const anyChecked = Array.from(checkboxes).some(checkbox => checkbox.checked);
    downloadBtn.disabled = !anyChecked;
}

async function handleDownload() {
    const checkboxes = tagsContainer.querySelectorAll('input[type="checkbox"]:checked');
    const selectedTags = Array.from(checkboxes).map(checkbox => checkbox.value);

    if (selectedTags.length === 0) {
        showMessage(downloadMessage, '请至少选择一个模块', 'error');
        return;
    }

    downloadBtn.disabled = true;
    downloadBtn.classList.add('loading');
    downloadBtn.textContent = '下载中...';
    hideMessage(downloadMessage);

    try {
        const response = await fetch('/api/filter', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ selectedTags }),
        });

        const data = await response.json();

        if (data.success && data.data) {
            // 下载过滤后的 JSON
            downloadJSON(data.data);
            showMessage(downloadMessage, `已成功下载包含 ${selectedTags.length} 个模块的 OpenAPI 文档！`, 'success');
        } else {
            showMessage(downloadMessage, data.message || '过滤 OpenAPI 失败', 'error');
        }
    } catch (error) {
        showMessage(downloadMessage, `错误：${error.message}`, 'error');
    } finally {
        downloadBtn.disabled = false;
        downloadBtn.classList.remove('loading');
        downloadBtn.textContent = '下载过滤后的 JSON';
    }
}

function downloadJSON(jsonString) {
    try {
        // 验证 JSON 格式
        JSON.parse(jsonString);

        const element = document.createElement('a');
        element.setAttribute('href', 'data:application/json;charset=utf-8,' + encodeURIComponent(jsonString));
        element.setAttribute('download', `openapi-filtered-${new Date().getTime()}.json`);
        element.style.display = 'none';
        document.body.appendChild(element);
        element.click();
        document.body.removeChild(element);
    } catch (error) {
        showMessage(downloadMessage, '错误：无效的 JSON 响应', 'error');
    }
}

// 页面加载时初始化
document.addEventListener('DOMContentLoaded', () => {
    updateDownloadButton();
});
