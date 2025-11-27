// 应用程序主要JavaScript文件
document.addEventListener('DOMContentLoaded', function() {
    // 初始化所有组件
    initTooltips();
    initAlerts();
    initForms();
    initTables();
    initModals();
    
    console.log('Hertz MVC Framework 前端初始化完成');
});

// 初始化工具提示
function initTooltips() {
    var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    var tooltipList = tooltipTriggerList.map(function (tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl);
    });
}

// 初始化警告框自动关闭
function initAlerts() {
    // 自动关闭警告框
    setTimeout(function() {
        var alerts = document.querySelectorAll('.alert:not(.alert-permanent)');
        alerts.forEach(function(alert) {
            var bsAlert = new bootstrap.Alert(alert);
            bsAlert.close();
        });
    }, 5000);
}

// 初始化表单验证
function initForms() {
    // Bootstrap 表单验证
    var forms = document.querySelectorAll('.needs-validation');
    Array.prototype.slice.call(forms).forEach(function(form) {
        form.addEventListener('submit', function(event) {
            if (!form.checkValidity()) {
                event.preventDefault();
                event.stopPropagation();
            }
            form.classList.add('was-validated');
        }, false);
    });
    
    // AJAX 表单提交
    var ajaxForms = document.querySelectorAll('.ajax-form');
    ajaxForms.forEach(function(form) {
        form.addEventListener('submit', function(e) {
            e.preventDefault();
            submitAjaxForm(form);
        });
    });
}

// 初始化表格功能
function initTables() {
    // 表格行点击事件
    var tableRows = document.querySelectorAll('.table tbody tr[data-href]');
    tableRows.forEach(function(row) {
        row.style.cursor = 'pointer';
        row.addEventListener('click', function() {
            window.location.href = this.dataset.href;
        });
    });
    
    // 全选功能
    var selectAllCheckbox = document.querySelector('#selectAll');
    if (selectAllCheckbox) {
        selectAllCheckbox.addEventListener('change', function() {
            var checkboxes = document.querySelectorAll('.row-checkbox');
            checkboxes.forEach(function(checkbox) {
                checkbox.checked = selectAllCheckbox.checked;
            });
        });
    }
}

// 初始化模态框
function initModals() {
    // 模态框显示时聚焦第一个输入框
    var modals = document.querySelectorAll('.modal');
    modals.forEach(function(modal) {
        modal.addEventListener('shown.bs.modal', function() {
            var firstInput = modal.querySelector('input, textarea, select');
            if (firstInput) {
                firstInput.focus();
            }
        });
        
        // 模态框关闭时清空表单
        modal.addEventListener('hidden.bs.modal', function() {
            var forms = modal.querySelectorAll('form');
            forms.forEach(function(form) {
                form.reset();
                form.classList.remove('was-validated');
            });
        });
    });
}

// AJAX表单提交
function submitAjaxForm(form) {
    var formData = new FormData(form);
    var submitBtn = form.querySelector('[type="submit"]');
    var originalText = submitBtn.textContent;
    
    // 显示加载状态
    submitBtn.disabled = true;
    submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm me-2" role="status"></span>提交中...';
    
    fetch(form.action, {
        method: form.method,
        body: formData,
        headers: {
            'X-Requested-With': 'XMLHttpRequest'
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showAlert('success', data.message || '操作成功');
            // 如果有回调函数，执行它
            if (form.dataset.onSuccess) {
                window[form.dataset.onSuccess](data);
            }
            // 关闭模态框
            var modal = form.closest('.modal');
            if (modal) {
                bootstrap.Modal.getInstance(modal).hide();
            }
        } else {
            showAlert('danger', data.message || '操作失败');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showAlert('danger', '网络错误，请稍后重试');
    })
    .finally(() => {
        // 恢复按钮状态
        submitBtn.disabled = false;
        submitBtn.textContent = originalText;
    });
}

// 显示警告框
function showAlert(type, message, duration = 5000) {
    var alertsContainer = document.getElementById('alerts-container');
    if (!alertsContainer) {
        alertsContainer = document.createElement('div');
        alertsContainer.id = 'alerts-container';
        alertsContainer.className = 'position-fixed top-0 end-0 p-3';
        alertsContainer.style.zIndex = '1055';
        document.body.appendChild(alertsContainer);
    }
    
    var alertId = 'alert-' + Date.now();
    var alertHTML = `
        <div class="alert alert-${type} alert-dismissible fade show" role="alert" id="${alertId}">
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        </div>
    `;
    
    alertsContainer.insertAdjacentHTML('beforeend', alertHTML);
    
    // 自动关闭
    if (duration > 0) {
        setTimeout(() => {
            var alert = document.getElementById(alertId);
            if (alert) {
                var bsAlert = new bootstrap.Alert(alert);
                bsAlert.close();
            }
        }, duration);
    }
}

// 确认对话框
function confirmAction(message, callback) {
    if (confirm(message)) {
        callback();
    }
}

// 删除项目
function deleteItem(url, id, callback) {
    confirmAction('确定要删除这个项目吗？此操作不可撤销。', function() {
        fetch(url + '?id=' + id, {
            method: 'DELETE',
            headers: {
                'X-Requested-With': 'XMLHttpRequest'
            }
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                showAlert('success', data.message || '删除成功');
                if (callback) callback(data);
            } else {
                showAlert('danger', data.message || '删除失败');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            showAlert('danger', '网络错误，请稍后重试');
        });
    });
}

// 加载数据
function loadData(url, containerId, showLoading = true) {
    var container = document.getElementById(containerId);
    if (!container) return;
    
    if (showLoading) {
        container.innerHTML = '<div class="text-center p-3"><div class="spinner-border" role="status"></div></div>';
    }
    
    fetch(url, {
        headers: {
            'X-Requested-With': 'XMLHttpRequest'
        }
    })
    .then(response => response.text())
    .then(html => {
        container.innerHTML = html;
        // 重新初始化组件
        initTooltips();
        initTables();
    })
    .catch(error => {
        console.error('Error:', error);
        container.innerHTML = '<div class="alert alert-danger">加载数据失败</div>';
    });
}

// 分页功能
function changePage(url, page) {
    var separator = url.includes('?') ? '&' : '?';
    window.location.href = url + separator + 'page=' + page;
}

// 搜索功能
function initSearch() {
    var searchInput = document.getElementById('searchInput');
    if (searchInput) {
        var searchTimeout;
        searchInput.addEventListener('input', function() {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                performSearch(this.value);
            }, 500);
        });
    }
}

function performSearch(query) {
    var currentUrl = new URL(window.location);
    if (query) {
        currentUrl.searchParams.set('search', query);
    } else {
        currentUrl.searchParams.delete('search');
    }
    currentUrl.searchParams.delete('page'); // 重置页码
    window.location.href = currentUrl.toString();
}

// 文件上传
function initFileUpload() {
    var uploadAreas = document.querySelectorAll('.upload-area');
    uploadAreas.forEach(function(area) {
        var input = area.querySelector('input[type="file"]');
        
        // 点击上传
        area.addEventListener('click', function() {
            input.click();
        });
        
        // 拖拽上传
        area.addEventListener('dragover', function(e) {
            e.preventDefault();
            area.classList.add('dragover');
        });
        
        area.addEventListener('dragleave', function() {
            area.classList.remove('dragover');
        });
        
        area.addEventListener('drop', function(e) {
            e.preventDefault();
            area.classList.remove('dragover');
            
            var files = e.dataTransfer.files;
            if (files.length > 0) {
                input.files = files;
                handleFileUpload(input);
            }
        });
        
        // 文件选择
        input.addEventListener('change', function() {
            handleFileUpload(this);
        });
    });
}

function handleFileUpload(input) {
    var files = input.files;
    if (files.length === 0) return;
    
    var formData = new FormData();
    for (var i = 0; i < files.length; i++) {
        formData.append('files[]', files[i]);
    }
    
    var uploadUrl = input.dataset.uploadUrl || '/upload';
    
    fetch(uploadUrl, {
        method: 'POST',
        body: formData,
        headers: {
            'X-Requested-With': 'XMLHttpRequest'
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showAlert('success', '文件上传成功');
            if (data.files) {
                displayUploadedFiles(data.files);
            }
        } else {
            showAlert('danger', data.message || '文件上传失败');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showAlert('danger', '上传过程中发生错误');
    });
}

// 实用工具函数
const Utils = {
    // 格式化日期
    formatDate: function(date, format = 'YYYY-MM-DD HH:mm:ss') {
        if (!(date instanceof Date)) {
            date = new Date(date);
        }
        
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        const seconds = String(date.getSeconds()).padStart(2, '0');
        
        return format
            .replace('YYYY', year)
            .replace('MM', month)
            .replace('DD', day)
            .replace('HH', hours)
            .replace('mm', minutes)
            .replace('ss', seconds);
    },
    
    // 格式化文件大小
    formatFileSize: function(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    },
    
    // 防抖函数
    debounce: function(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    },
    
    // 节流函数
    throttle: function(func, limit) {
        let inThrottle;
        return function() {
            const args = arguments;
            const context = this;
            if (!inThrottle) {
                func.apply(context, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }
};

// 全局错误处理
window.addEventListener('error', function(e) {
    console.error('全局错误:', e.error);
    // 可以在这里添加错误上报逻辑
});

// 页面可见性变化处理
document.addEventListener('visibilitychange', function() {
    if (document.hidden) {
        console.log('页面已隐藏');
    } else {
        console.log('页面已显示');
        // 可以在这里添加页面重新显示时的逻辑
    }
});