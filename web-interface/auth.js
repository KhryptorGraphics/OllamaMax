/**
 * Authentication UI Logic
 * Handles login, registration, and password validation
 */

const API_BASE = 'http://localhost:13100';

// Tab switching
document.querySelectorAll('.auth-tab').forEach(tab => {
    tab.addEventListener('click', () => {
        const tabName = tab.dataset.tab;
        
        // Update active tab
        document.querySelectorAll('.auth-tab').forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        
        // Update active form
        document.querySelectorAll('.auth-form').forEach(f => f.classList.remove('active'));
        document.getElementById(`${tabName}Form`).classList.add('active');
        
        // Clear messages
        hideMessages();
    });
});

// Password strength checker
const passwordInput = document.getElementById('registerPassword');
const strengthBar = document.getElementById('passwordStrengthBar');

const requirements = {
    length: { regex: /.{8,}/, element: document.getElementById('req-length') },
    uppercase: { regex: /[A-Z]/, element: document.getElementById('req-uppercase') },
    lowercase: { regex: /[a-z]/, element: document.getElementById('req-lowercase') },
    number: { regex: /[0-9]/, element: document.getElementById('req-number') },
    special: { regex: /[@$!%*?&]/, element: document.getElementById('req-special') }
};

passwordInput.addEventListener('input', (e) => {
    const password = e.target.value;
    let metCount = 0;
    
    // Check each requirement
    Object.keys(requirements).forEach(key => {
        const req = requirements[key];
        const met = req.regex.test(password);
        
        if (met) {
            req.element.classList.add('met');
            req.element.querySelector('.icon').textContent = '✓';
            metCount++;
        } else {
            req.element.classList.remove('met');
            req.element.querySelector('.icon').textContent = '○';
        }
    });
    
    // Update strength bar
    strengthBar.className = 'password-strength-bar';
    if (metCount === 0) {
        strengthBar.className = 'password-strength-bar';
    } else if (metCount <= 2) {
        strengthBar.classList.add('weak');
    } else if (metCount <= 4) {
        strengthBar.classList.add('medium');
    } else {
        strengthBar.classList.add('strong');
    }
    
    // Update input border
    if (metCount === 5) {
        passwordInput.classList.remove('error');
        passwordInput.classList.add('success');
    } else if (password.length > 0) {
        passwordInput.classList.remove('success');
        passwordInput.classList.add('error');
    } else {
        passwordInput.classList.remove('error', 'success');
    }
});

// Confirm password validation
const confirmPasswordInput = document.getElementById('registerConfirmPassword');
confirmPasswordInput.addEventListener('input', (e) => {
    const password = passwordInput.value;
    const confirmPassword = e.target.value;
    
    if (confirmPassword.length > 0) {
        if (password === confirmPassword) {
            confirmPasswordInput.classList.remove('error');
            confirmPasswordInput.classList.add('success');
        } else {
            confirmPasswordInput.classList.remove('success');
            confirmPasswordInput.classList.add('error');
        }
    } else {
        confirmPasswordInput.classList.remove('error', 'success');
    }
});

// Login form submission
document.getElementById('loginForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const email = document.getElementById('loginEmail').value;
    const password = document.getElementById('loginPassword').value;
    const btn = e.target.querySelector('.submit-btn');
    
    setLoading(btn, true);
    hideMessages();
    
    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            const data = await response.json();

            showSuccess('Login successful! Redirecting...');
            setTimeout(() => {
                window.location.href = 'index.html';
            }, 1500);
        } else {
            showError(data.message || 'Login failed. Please check your credentials.');
        }
    } catch (error) {
        showError('Network error. Please check your connection.');
    } finally {
        setLoading(btn, false);
    }
});

// Register form submission
document.getElementById('registerForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const firstName = document.getElementById('registerFirstName').value;
    const lastName = document.getElementById('registerLastName').value;
    const email = document.getElementById('registerEmail').value;
    const password = document.getElementById('registerPassword').value;
    const confirmPassword = document.getElementById('registerConfirmPassword').value;
    const btn = e.target.querySelector('.submit-btn');
    
    hideMessages();
    
    // Validate password requirements
    const allMet = Object.keys(requirements).every(key => 
        requirements[key].regex.test(password)
    );
    
    if (!allMet) {
        showError('Please meet all password requirements.');
        return;
    }
    
    if (password !== confirmPassword) {
        showError('Passwords do not match.');
        return;
    }

    setLoading(btn, true);

    try {
        const response = await fetch(`${API_BASE}/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                firstName,
                lastName,
                email,
                password,
                confirmPassword
            })
        });

        const data = await response.json();

        if (response.ok) {
            showSuccess('Registration successful! Please check your email for verification.');
            // Clear form
            e.target.reset();
            // Reset password requirements
            Object.keys(requirements).forEach(key => {
                requirements[key].element.classList.remove('met');
                requirements[key].element.querySelector('.icon').textContent = '○';
            });
            strengthBar.className = 'password-strength-bar';
        } else {
            showError(data.message || 'Registration failed. Please try again.');
        }
    } catch (error) {
        showError('Network error. Please check your connection.');
    } finally {
        setLoading(btn, false);
    }
});

// Helper functions
function showError(message) {
    const errorEl = document.getElementById('errorMessage');
    errorEl.textContent = message;
    errorEl.classList.add('show');
}

function showSuccess(message) {
    const successEl = document.getElementById('successMessage');
    successEl.textContent = message;
    successEl.classList.add('show');
}

function hideMessages() {
    document.getElementById('errorMessage').classList.remove('show');
    document.getElementById('successMessage').classList.remove('show');
}

function setLoading(button, loading) {
    const btnText = button.querySelector('.btn-text');

    if (loading) {
        button.disabled = true;
        btnText.innerHTML = '<span class="loading"></span>';
    } else {
        button.disabled = false;
        const form = button.closest('form');
        if (form.id === 'loginForm') {
            btnText.textContent = 'Login';
        } else {
            btnText.textContent = 'Create Account';
        }
    }
}

