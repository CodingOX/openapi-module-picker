// DOM Elements
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

// Initialize theme
initializeTheme();

// Event listeners
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

async function handleParse() {
    const url = openApiUrlInput.value.trim();
    
    if (!url) {
        showMessage(parseMessage, 'Please enter a valid URL', 'error');
        return;
    }

    parseBtn.disabled = true;
    parseBtn.classList.add('loading');
    parseBtn.textContent = 'Loading...';
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
            infoBar.innerHTML = `<strong>OpenAPI ${data.version}</strong> loaded successfully with <strong>${allTags.length}</strong> module(s)`;
            showMessage(parseMessage, 'OpenAPI parsed successfully!', 'success');
        } else {
            showMessage(parseMessage, data.message || 'Failed to parse OpenAPI', 'error');
        }
    } catch (error) {
        showMessage(parseMessage, `Error: ${error.message}`, 'error');
    } finally {
        parseBtn.disabled = false;
        parseBtn.classList.remove('loading');
        parseBtn.textContent = 'Load & Parse';
    }
}

function renderTags() {
    tagsContainer.innerHTML = '';
    
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
        showMessage(downloadMessage, 'Please select at least one module', 'error');
        return;
    }

    downloadBtn.disabled = true;
    downloadBtn.classList.add('loading');
    downloadBtn.textContent = 'Downloading...';
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
            // Download the filtered JSON
            downloadJSON(data.data);
            showMessage(downloadMessage, `Successfully downloaded filtered OpenAPI with ${selectedTags.length} module(s)!`, 'success');
        } else {
            showMessage(downloadMessage, data.message || 'Failed to filter OpenAPI', 'error');
        }
    } catch (error) {
        showMessage(downloadMessage, `Error: ${error.message}`, 'error');
    } finally {
        downloadBtn.disabled = false;
        downloadBtn.classList.remove('loading');
        downloadBtn.textContent = 'Download Filtered JSON';
    }
}

function downloadJSON(jsonString) {
    try {
        // Validate JSON
        JSON.parse(jsonString);
        
        const element = document.createElement('a');
        element.setAttribute('href', 'data:application/json;charset=utf-8,' + encodeURIComponent(jsonString));
        element.setAttribute('download', `openapi-filtered-${new Date().getTime()}.json`);
        element.style.display = 'none';
        document.body.appendChild(element);
        element.click();
        document.body.removeChild(element);
    } catch (error) {
        showMessage(downloadMessage, 'Error: Invalid JSON response', 'error');
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    updateDownloadButton();
});
