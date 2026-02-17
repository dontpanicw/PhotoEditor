const API_BASE = window.location.origin;

let uploadedImages = [];
let statusCheckIntervals = {};

document.addEventListener('DOMContentLoaded', () => {
    console.log('App loaded, version 2');
    loadImagesFromStorage();
    
    // Очищаем изображения без ID
    uploadedImages = uploadedImages.filter(img => img.id && img.id.trim() !== '');
    saveImagesToStorage();
    
    renderImages();
    document.getElementById('uploadForm').addEventListener('submit', handleUpload);
    
    // Проверяем статусы всех pending изображений
    uploadedImages.forEach(img => {
        if (img.status === 'Pending' && img.id) {
            console.log('Checking status for pending image:', img.id);
            checkImageStatus(img.id);
        }
    });
});

async function handleUpload(e) {
    e.preventDefault();
    
    const fileInput = document.getElementById('imageInput');
    const file = fileInput.files[0];
    
    if (!file) {
        showStatus('Выберите файл', 'error');
        return;
    }
    
    if (!file.type.startsWith('image/')) {
        showStatus('Пожалуйста, выберите изображение', 'error');
        return;
    }
    
    // Собираем выбранные действия
    const checkboxes = document.querySelectorAll('input[name="action"]:checked');
    const actions = Array.from(checkboxes).map(cb => cb.value);
    
    if (actions.length === 0) {
        showStatus('Выберите хотя бы одно действие', 'error');
        return;
    }
    
    const formData = new FormData();
    formData.append('image', file);
    formData.append('actions', actions.join(','));
    
    const submitBtn = e.target.querySelector('button[type="submit"]');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Загрузка...';
    
    try {
        const response = await fetch(`${API_BASE}/upload`, {
            method: 'POST',
            body: formData
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(errorText || 'Ошибка загрузки');
        }
        
        const data = await response.json();

        
        const imageData = {
            id: data.id,
            status: 'Pending',
            filename: file.name,
            actions: actions,
            uploadedAt: new Date().toISOString()
        };
        
        uploadedImages.unshift(imageData);
        saveImagesToStorage();
        renderImages();
        
        showStatus(`✅ Изображение загружено! Применяются действия: ${actions.join(', ')}`, 'success');
        fileInput.value = '';
        
        checkImageStatus(data.id);
        
    } catch (error) {
        console.error('Upload error:', error);
        showStatus('❌ Ошибка: ' + error.message, 'error');
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = 'Загрузить и обработать';
    }
}

function checkImageStatus(imageId) {
    if (!imageId || imageId.trim() === '') {
        console.error('Invalid imageId:', imageId);
        return;
    }
    
    if (statusCheckIntervals[imageId]) {
        console.log('Status check already running for:', imageId);
        return;
    }
    
    let attempts = 0;
    const maxAttempts = 60;
    
    console.log('Starting status check for:', imageId);
    
    statusCheckIntervals[imageId] = setInterval(async () => {
        attempts++;
        
        try {
            // Используем новый endpoint для проверки статуса
            const url = `${API_BASE}/image/${imageId}/status`;
            console.log(`Checking status (attempt ${attempts}):`, url);
            
            const response = await fetch(url);
            
            if (response.ok) {
                const data = await response.json();
                console.log('Status response:', data);
                
                if (data.status === 'Done') {
                    console.log('Image is ready:', imageId);
                    clearInterval(statusCheckIntervals[imageId]);
                    delete statusCheckIntervals[imageId];
                    updateImageStatus(imageId, 'Done');
                    showStatus('✅ Изображение готово!', 'success');
                } else if (data.status === 'Failed') {
                    console.log('Image processing failed:', imageId);
                    clearInterval(statusCheckIntervals[imageId]);
                    delete statusCheckIntervals[imageId];
                    updateImageStatus(imageId, 'Failed');
                    showStatus('❌ Ошибка обработки изображения', 'error');
                } else {
                    console.log('Image still pending:', imageId, 'status:', data.status);
                }
                
                if (attempts >= maxAttempts) {
                    console.log('Max attempts reached for:', imageId);
                    clearInterval(statusCheckIntervals[imageId]);
                    delete statusCheckIntervals[imageId];
                    updateImageStatus(imageId, 'Failed');
                    showStatus('❌ Превышено время ожидания обработки', 'error');
                }
            } else {
                console.error('Status check failed:', response.status, response.statusText);
                if (attempts >= maxAttempts) {
                    clearInterval(statusCheckIntervals[imageId]);
                    delete statusCheckIntervals[imageId];
                    updateImageStatus(imageId, 'Failed');
                    showStatus('❌ Превышено время ожидания обработки', 'error');
                }
            }
        } catch (error) {
            console.error('Status check error:', error);
            if (attempts >= maxAttempts) {
                clearInterval(statusCheckIntervals[imageId]);
                delete statusCheckIntervals[imageId];
                updateImageStatus(imageId, 'Failed');
            }
        }
    }, 3000);
}

function updateImageStatus(imageId, status) {
    const image = uploadedImages.find(img => img.id === imageId);
    if (image) {
        image.status = status;
        saveImagesToStorage();
        renderImages();
    }
}

async function viewImage(imageId) {
    try {
        const url = `${API_BASE}/image/${imageId}`;
        window.open(url, '_blank');
    } catch (error) {
        showStatus('❌ Ошибка: ' + error.message, 'error');
    }
}

async function deleteImage(imageId) {
    if (!confirm('Удалить это изображение?')) {
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/image/${imageId}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            throw new Error('Ошибка удаления');
        }
        
        if (statusCheckIntervals[imageId]) {
            clearInterval(statusCheckIntervals[imageId]);
            delete statusCheckIntervals[imageId];
        }
        
        uploadedImages = uploadedImages.filter(img => img.id !== imageId);
        saveImagesToStorage();
        renderImages();
        
        showStatus('✅ Изображение удалено', 'success');
        
    } catch (error) {
        console.error('Delete error:', error);
        showStatus('❌ Ошибка: ' + error.message, 'error');
    }
}

function renderImages() {
    const container = document.getElementById('imagesList');
    
    if (uploadedImages.length === 0) {
        container.innerHTML = '<div class="empty-state">📭 Нет загруженных изображений</div>';
        return;
    }
    
    container.innerHTML = uploadedImages.map(image => {
        const statusText = getStatusText(image.status);
        const statusIcon = getStatusIcon(image.status);
        
        return `
            <div class="image-card">
                ${image.status === 'Done' 
                    ? `<img src="${API_BASE}/image/${image.id}" class="image-preview" alt="${image.filename}" 
                           onerror="this.parentElement.innerHTML='<div class=\\'image-placeholder\\'>❌ Ошибка загрузки</div>'">` 
                    : `<div class="image-placeholder">${statusIcon} ${statusText}</div>`
                }
                <div class="image-info">
                    <div class="image-filename" title="${image.filename}">📄 ${truncateFilename(image.filename)}</div>
                    <div class="image-id">🆔 ${image.id}</div>
                    <span class="image-status ${image.status.toLowerCase()}">${statusIcon} ${statusText}</span>
                </div>
                <div class="image-actions">
                    <button class="btn-view" onclick="viewImage('${image.id}')" ${image.status !== 'Done' ? 'disabled' : ''}>
                        👁️ Просмотр
                    </button>
                    <button class="btn-delete" onclick="deleteImage('${image.id}')">
                        🗑️ Удалить
                    </button>
                </div>
            </div>
        `;
    }).join('');
}

function getStatusText(status) {
    const statusMap = {
        'Pending': 'Обработка',
        'Done': 'Готово',
        'Failed': 'Ошибка'
    };
    return statusMap[status] || status;
}

function getStatusIcon(status) {
    const iconMap = {
        'Pending': '⏳',
        'Done': '✅',
        'Failed': '❌'
    };
    return iconMap[status] || '❓';
}

function truncateFilename(filename, maxLength = 30) {
    if (filename.length <= maxLength) return filename;
    const ext = filename.split('.').pop();
    const name = filename.substring(0, filename.lastIndexOf('.'));
    const truncated = name.substring(0, maxLength - ext.length - 4) + '...';
    return truncated + '.' + ext;
}

function showStatus(message, type) {
    const statusDiv = document.getElementById('uploadStatus');
    statusDiv.textContent = message;
    statusDiv.className = `status ${type}`;
    statusDiv.style.display = 'block';
    
    setTimeout(() => {
        statusDiv.style.display = 'none';
    }, 5000);
}

function saveImagesToStorage() {
    try {
        localStorage.setItem('uploadedImages', JSON.stringify(uploadedImages));
    } catch (e) {
        console.error('Failed to save to localStorage:', e);
    }
}

function loadImagesFromStorage() {
    try {
        const stored = localStorage.getItem('uploadedImages');
        if (stored) {
            uploadedImages = JSON.parse(stored);
        }
    } catch (e) {
        console.error('Failed to load from localStorage:', e);
        uploadedImages = [];
    }
}
