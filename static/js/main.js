// Show logout button on index if logged in
if (window.location.pathname.endsWith('index.html') || window.location.pathname === '/' ) {
  if (localStorage.getItem('token')) {
    document.getElementById('logoutBtnIndex').classList.remove('d-none');
  }
}

document.getElementById('logoutBtnIndex')?.addEventListener('click', () => {
  localStorage.removeItem('token');
  window.location.href = 'index.html';
});

function showMessage(msg, type = 'info') {
  const msgDiv = document.getElementById('message');
  msgDiv.textContent = msg;
  msgDiv.className = `alert alert-${type}`;
  msgDiv.classList.remove('d-none');
}

// Redirect to dashboard if already logged in
if (window.location.pathname.endsWith('index.html') || window.location.pathname === '/' ) {
  if (localStorage.getItem('token')) {
    // Already handled above for logout button
  }
}

// Redirect to login if not logged in
if (window.location.pathname.endsWith('dashboard.html')) {
  if (!localStorage.getItem('token')) {
    window.location.href = 'index.html';
  }
}

// Helper: Check if profile is complete
async function checkProfileComplete(token) {
  const res = await fetch('/profile', {
    method: 'GET',
    headers: { 'Authorization': 'Bearer ' + token }
  });
  return res.ok;
}

// After login, check profile completeness before redirecting
async function handlePostLogin(token) {
  localStorage.setItem('token', token);
  const complete = await checkProfileComplete(token);
  if (complete) {
    window.location.href = 'dashboard.html';
  } else {
    window.location.href = 'profile.html';
  }
}

// Handle login form
document.getElementById('loginForm')?.addEventListener('submit', async (e) => {
  e.preventDefault();
  const email = document.getElementById('loginEmail').value;
  const otp = document.getElementById('loginOTP').value;
  let res, data;
  try {
    res = await fetch('/login', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({ email, code: otp })
    });
    data = await res.json();
  } catch (err) {
    showMessage('Network error. Please try again.', 'danger');
    return;
  }
  if (res.ok && data.token) {
    handlePostLogin(data.token);
  } else {
    showMessage(data.error || 'Login failed. Make sure you have registered and entered the correct OTP.', 'danger');
  }
});

// Handle register form
document.getElementById('registerForm')?.addEventListener('submit', async (e) => {
  e.preventDefault();
  const email = document.getElementById('registerEmail').value;
  let res, data;
  try {
    res = await fetch('/register', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({ email })
    });
    if (res.ok) {
      showMessage('OTP sent to your email! Please check your inbox and enter the OTP in the Login tab.', 'success');
    } else {
      data = await res.json();
      showMessage(data.error || 'Registration failed', 'danger');
    }
  } catch (err) {
    showMessage('Network error. Please try again.', 'danger');
  }
});

// Google OAuth
document.getElementById('googleLoginBtn')?.addEventListener('click', () => {
  window.location.href = '/auth/google';
});
document.getElementById('googleRegisterBtn')?.addEventListener('click', () => {
  window.location.href = '/auth/google';
});

// Logout on dashboard
document.getElementById('logoutBtn')?.addEventListener('click', () => {
  localStorage.removeItem('token');
  window.location.href = 'index.html';
});

// Profile completion logic for profile.html
if (window.location.pathname.endsWith('profile.html')) {
  // Branch options
  const majors = ["Math", "Chemistry", "Physics", "Economics", "Biology"];
  const independents = ["CS", "ECE", "EEE", "ENI", "Mech", "Chemical", "Manu.", "Civil", "B.pharma"];
  const minors = ["Civil", "Manu.", "Chemical", "Mech", "ENI", "EEE", "ECE", "MNC", "CS"];
  const branchSelect = document.getElementById('profileBranch');
  let options = [];
  majors.forEach(major => {
    options.push(`<option value="${major}">${major}</option>`);
    minors.forEach(minor => {
      options.push(`<option value="${major}+${minor}">${major}+${minor}</option>`);
    });
  });
  independents.forEach(ind => {
    options.push(`<option value="${ind}">${ind}</option>`);
  });
  branchSelect.innerHTML = '<option value="">Select Branch</option>' + options.join('');

  // Get email from token (decode JWT)
  function parseJwt (token) {
    try {
      return JSON.parse(atob(token.split('.')[1]));
    } catch (e) {
      return null;
    }
  }
  const token = localStorage.getItem('token');
  let email = '';
  if (token) {
    const payload = parseJwt(token);
    if (payload && payload.email) email = payload.email;
  }

  // Detect batch from email
  let batch = '';
  let batchValid = false;
  const batchMatch = email.match(/^f(202[1-4])/);
  if (batchMatch) {
    batch = batchMatch[1];
    batchValid = true;
  }
  document.getElementById('profileBatch').value = batch ? batch + ' Batch' : '';

  // Show error if batch is not valid
  if (!batchValid) {
    document.getElementById('profileError').classList.remove('d-none');
    document.getElementById('profileError').textContent = 'Only 2021, 2022, 2023, 2024 batches are allowed.';
    document.getElementById('profileForm').querySelector('button[type="submit"]').disabled = true;
  }

  // Handle profile form submit
  document.getElementById('profileForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    if (!batchValid) return;
    const name = document.getElementById('profileName').value;
    const age = document.getElementById('profileAge').value;
    const branch = document.getElementById('profileBranch').value;
    // Save profile to backend
    const res = await fetch('/profile', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + token
      },
      body: JSON.stringify({ name, age, branch, batch })
    });
    if (res.ok) {
      window.location.href = 'dashboard.html';
    } else {
      document.getElementById('profileError').classList.remove('d-none');
      document.getElementById('profileError').textContent = 'Failed to save profile. Please try again.';
    }
  });
}

// On dashboard.html, check for token in URL fragment from Google OAuth
if (window.location.pathname.endsWith('dashboard.html')) {
  const hash = window.location.hash;
  if (hash.startsWith('#token=')) {
    const token = hash.replace('#token=', '');
    handlePostLogin(token);
    // Clean up the URL (remove fragment)
    window.location.hash = '';
    // Optionally reload to ensure clean state
    // window.location.reload();
  } else {
    // If not coming from Google, check profile completeness
    const token = localStorage.getItem('token');
    if (token) {
      checkProfileComplete(token).then(complete => {
        if (!complete) {
          window.location.href = 'profile.html';
        }
      });
    }
  }
} 