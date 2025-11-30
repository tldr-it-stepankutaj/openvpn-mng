// OpenVPN Manager JavaScript

// Logout function
async function logout() {
    try {
        await fetch('/api/v1/auth/logout', { method: 'POST' });
    } catch (error) {
        console.error('Logout error:', error);
    }
    globalThis.location.href = '/login';
}

// API helper functions
const api = {
    async get(url) {
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(await this.getErrorMessage(response));
        }
        return response.json();
    },

    async post(url, data) {
        const response = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        if (!response.ok) {
            throw new Error(await this.getErrorMessage(response));
        }
        return response.json();
    },

    async put(url, data) {
        const response = await fetch(url, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        if (!response.ok) {
            throw new Error(await this.getErrorMessage(response));
        }
        return response.json();
    },

    async delete(url) {
        const response = await fetch(url, { method: 'DELETE' });
        if (!response.ok) {
            throw new Error(await this.getErrorMessage(response));
        }
        return response.json();
    },

    async getErrorMessage(response) {
        try {
            const data = await response.json();
            return data.message || 'An error occurred';
        } catch {
            return 'An error occurred';
        }
    }
};

// Alert helper
function showAlert(containerId, message, type = 'danger') {
    const container = document.getElementById(containerId);
    if (container) {
        container.className = `alert alert-${type}`;
        container.textContent = message;
        container.classList.remove('d-none');
    }
}

function hideAlert(containerId) {
    const container = document.getElementById(containerId);
    if (container) {
        container.classList.add('d-none');
    }
}

// Confirmation dialog
function confirmAction(message) {
    return confirm(message);
}

// Format date
function formatDate(dateString) {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
}

// Initialize tooltips and popovers (Bootstrap)
document.addEventListener('DOMContentLoaded', function() {
    // Initialize Bootstrap tooltips
    const tooltipTriggerList = Array.prototype.slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    tooltipTriggerList.map(function(tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl);
    });

    // Initialize Bootstrap popovers
    const popoverTriggerList = Array.prototype.slice.call(document.querySelectorAll('[data-bs-toggle="popover"]'));
    popoverTriggerList.map(function(popoverTriggerEl) {
        return new bootstrap.Popover(popoverTriggerEl);
    });
});
