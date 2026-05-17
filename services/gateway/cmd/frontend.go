package main

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>GoTicket — Admin Dashboard</title>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:'Inter',sans-serif;background:#0f172a;color:#e2e8f0;display:flex;min-height:100vh}
.sidebar{width:240px;background:#1e293b;display:flex;flex-direction:column;padding:24px 0;border-right:1px solid #334155;flex-shrink:0}
.logo{padding:0 24px 32px;display:flex;align-items:center;gap:10px;font-size:18px;font-weight:700;color:#f1f5f9}
.logo span{color:#10b981}
.nav-item{padding:10px 24px;cursor:pointer;display:flex;align-items:center;gap:10px;color:#94a3b8;border-left:3px solid transparent;transition:all .2s;font-size:14px}
.nav-item:hover,.nav-item.active{background:#0f172a;color:#10b981;border-left-color:#10b981}
.nav-icon{font-size:16px}
.sidebar-bottom{margin-top:auto;padding:16px 24px;border-top:1px solid #334155}
.main{flex:1;overflow:auto;padding:32px}
.header{display:flex;justify-content:space-between;align-items:center;margin-bottom:32px}
.header h1{font-size:24px;font-weight:700;color:#f1f5f9}
.badge{background:#10b98120;color:#10b981;padding:4px 12px;border-radius:20px;font-size:12px}
.stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:16px;margin-bottom:32px}
.card{background:#1e293b;border:1px solid #334155;border-radius:12px;padding:20px;transition:border-color .2s}
.card:hover{border-color:#10b981}
.card-label{font-size:12px;color:#64748b;margin-bottom:8px;text-transform:uppercase;letter-spacing:.5px}
.card-val{font-size:28px;font-weight:700;color:#10b981}
.card-sub{font-size:12px;color:#94a3b8;margin-top:4px}
.section{background:#1e293b;border:1px solid #334155;border-radius:12px;padding:24px;margin-bottom:24px}
.section-title{font-size:16px;font-weight:600;margin-bottom:20px;color:#f1f5f9;display:flex;align-items:center;gap:8px}
.form-row{display:grid;grid-template-columns:1fr 1fr;gap:12px;margin-bottom:12px}
.form-row.single{grid-template-columns:1fr}
label{display:block;font-size:12px;color:#94a3b8;margin-bottom:4px}
input,select{width:100%;background:#0f172a;border:1px solid #334155;border-radius:8px;padding:10px 14px;color:#e2e8f0;font-size:14px;outline:none;transition:border-color .2s}
input:focus,select:focus{border-color:#10b981}
.btn{padding:10px 20px;border-radius:8px;font-size:14px;font-weight:500;cursor:pointer;border:none;transition:all .2s}
.btn-primary{background:#10b981;color:#fff}
.btn-primary:hover{background:#059669}
.btn-secondary{background:#334155;color:#e2e8f0}
.btn-secondary:hover{background:#475569}
.btn-danger{background:#ef444420;color:#ef4444;border:1px solid #ef444440}
.btn-danger:hover{background:#ef4444;color:#fff}
.btn-sm{padding:6px 14px;font-size:12px}
.alert{padding:12px 16px;border-radius:8px;font-size:14px;margin-bottom:16px;display:none}
.alert.success{background:#10b98120;border:1px solid #10b98140;color:#10b981}
.alert.error{background:#ef444420;border:1px solid #ef444440;color:#ef4444}
.alert.show{display:block}
table{width:100%;border-collapse:collapse;font-size:13px}
th{text-align:left;padding:10px 12px;background:#0f172a;color:#64748b;font-weight:500;font-size:11px;text-transform:uppercase;letter-spacing:.5px}
td{padding:10px 12px;border-top:1px solid #334155;color:#cbd5e1}
tr:hover td{background:#0f172a20}
.tag{display:inline-block;padding:2px 10px;border-radius:20px;font-size:11px;font-weight:500}
.tag-pending{background:#f59e0b20;color:#f59e0b}
.tag-active{background:#10b98120;color:#10b981}
.tag-completed{background:#3b82f620;color:#3b82f6}
.tag-cancelled{background:#ef444420;color:#ef4444}
.tag-refunded{background:#a855f720;color:#a855f7}
.token-box{background:#0f172a;border:1px solid #334155;border-radius:8px;padding:12px;font-size:11px;color:#64748b;word-break:break-all;margin-top:8px;display:none}
.token-box.show{display:block}
.page-tabs{display:none}
.page-tabs.active{display:block}
.actions{display:flex;gap:8px}
.empty{text-align:center;padding:40px;color:#475569;font-size:14px}
.monitoring-grid{display:grid;grid-template-columns:1fr 1fr;gap:16px}
.metric-row{display:flex;justify-content:space-between;align-items:center;padding:8px 0;border-bottom:1px solid #334155;font-size:13px}
.metric-row:last-child{border-bottom:none}
.metric-key{color:#94a3b8}
.metric-val{color:#10b981;font-weight:600}
.up{color:#10b981}
.loading{opacity:.5}
</style>
</head>
<body>
<div class="sidebar">
  <div class="logo">🎫 <span>Go</span>Ticket</div>
  <div class="nav-item active" onclick="showPage('dashboard')"><span class="nav-icon">📊</span> Dashboard</div>
  <div class="nav-item" onclick="showPage('auth')"><span class="nav-icon">🔐</span> Auth</div>
  <div class="nav-item" onclick="showPage('users')"><span class="nav-icon">👥</span> Users</div>
  <div class="nav-item" onclick="showPage('orders')"><span class="nav-icon">🛒</span> Orders</div>
  <div class="nav-item" onclick="showPage('payments')"><span class="nav-icon">💳</span> Payments</div>
  <div class="nav-item" onclick="showPage('monitoring')"><span class="nav-icon">📈</span> Monitoring</div>
  <div class="sidebar-bottom">
    <div id="authStatus" style="font-size:12px;color:#64748b">Not logged in</div>
    <button class="btn btn-danger btn-sm" style="margin-top:8px;width:100%" onclick="doLogout()">Logout</button>
  </div>
</div>

<div class="main">
  <!-- DASHBOARD -->
  <div id="page-dashboard" class="page-tabs active">
    <div class="header"><h1>Dashboard</h1><span class="badge">GoTicket v1.0</span></div>
    <div class="stats">
      <div class="card"><div class="card-label">Total Users</div><div class="card-val" id="stat-users">—</div><div class="card-sub">Registered accounts</div></div>
      <div class="card"><div class="card-label">Total Orders</div><div class="card-val" id="stat-orders">—</div><div class="card-sub">All time</div></div>
      <div class="card"><div class="card-label">Payments</div><div class="card-val" id="stat-payments">—</div><div class="card-sub">Transactions</div></div>
      <div class="card"><div class="card-label">Gateway</div><div class="card-val up">✓ UP</div><div class="card-sub">API is running</div></div>
    </div>
    <div class="section">
      <div class="section-title">🚀 Quick Start</div>
      <p style="font-size:14px;color:#94a3b8;line-height:1.8">
        1. Go to <b style="color:#10b981">Auth</b> → Register a new user<br>
        2. Login and copy your <b style="color:#10b981">Access Token</b><br>
        3. Go to <b style="color:#10b981">Orders</b> → Create an order<br>
        4. Check <b style="color:#10b981">Payments</b> — payment is created automatically via NATS<br>
        5. View <b style="color:#10b981">Monitoring</b> for live metrics
      </p>
    </div>
  </div>

  <!-- AUTH -->
  <div id="page-auth" class="page-tabs">
    <div class="header"><h1>Auth</h1></div>
    <div style="display:grid;grid-template-columns:1fr 1fr;gap:24px">
      <div class="section">
        <div class="section-title">📝 Register</div>
        <div id="reg-alert" class="alert"></div>
        <div class="form-row single"><label>Name</label><input id="reg-name" placeholder="John Doe"></div>
        <div class="form-row single"><label>Email</label><input id="reg-email" type="email" placeholder="john@example.com"></div>
        <div class="form-row single"><label>Password</label><input id="reg-password" type="password" placeholder="Min 6 chars"></div>
        <button class="btn btn-primary" style="margin-top:8px" onclick="doRegister()">Register</button>
        <div id="reg-token" class="token-box"></div>
      </div>
      <div class="section">
        <div class="section-title">🔑 Login</div>
        <div id="login-alert" class="alert"></div>
        <div class="form-row single"><label>Email</label><input id="login-email" placeholder="john@example.com"></div>
        <div class="form-row single"><label>Password</label><input id="login-password" type="password" placeholder="Password"></div>
        <button class="btn btn-primary" style="margin-top:8px" onclick="doLogin()">Login</button>
        <div id="login-token" class="token-box"></div>
      </div>
    </div>
  </div>

  <!-- USERS -->
  <div id="page-users" class="page-tabs">
    <div class="header"><h1>Users</h1><button class="btn btn-primary btn-sm" onclick="loadUsers()">↻ Refresh</button></div>
    <div class="section">
      <div id="users-alert" class="alert"></div>
      <table><thead><tr><th>ID</th><th>Name</th><th>Email</th><th>Banned</th><th>Created</th><th>Actions</th></tr></thead>
      <tbody id="users-table"><tr><td colspan="6" class="empty">Click Refresh to load users</td></tr></tbody></table>
    </div>
  </div>

  <!-- ORDERS -->
  <div id="page-orders" class="page-tabs">
    <div class="header"><h1>Orders</h1><button class="btn btn-primary btn-sm" onclick="loadOrders()">↻ Refresh</button></div>
    <div class="section" style="margin-bottom:16px">
      <div class="section-title">➕ Create Order</div>
      <div id="order-alert" class="alert"></div>
      <div class="form-row">
        <div><label>User ID</label><input id="o-userid" type="number" placeholder="1"></div>
        <div><label>Product Name</label><input id="o-product" placeholder="Ticket VIP"></div>
      </div>
      <div class="form-row">
        <div><label>Quantity</label><input id="o-qty" type="number" value="1"></div>
        <div><label>Price (KZT)</label><input id="o-price" type="number" value="5000"></div>
      </div>
      <button class="btn btn-primary" onclick="createOrder()">Create Order</button>
    </div>
    <div class="section">
      <table><thead><tr><th>ID</th><th>User ID</th><th>Status</th><th>Total</th><th>Items</th><th>Created</th><th>Actions</th></tr></thead>
      <tbody id="orders-table"><tr><td colspan="7" class="empty">Click Refresh to load orders</td></tr></tbody></table>
    </div>
  </div>

  <!-- PAYMENTS -->
  <div id="page-payments" class="page-tabs">
    <div class="header"><h1>Payments</h1><button class="btn btn-primary btn-sm" onclick="loadPayments()">↻ Refresh</button></div>
    <div class="section">
      <div id="pay-alert" class="alert"></div>
      <table><thead><tr><th>ID</th><th>Order ID</th><th>User ID</th><th>Amount</th><th>Currency</th><th>Status</th><th>Method</th><th>Created</th></tr></thead>
      <tbody id="payments-table"><tr><td colspan="8" class="empty">Click Refresh to load payments</td></tr></tbody></table>
    </div>
  </div>

  <!-- MONITORING -->
  <div id="page-monitoring" class="page-tabs">
    <div class="header"><h1>Monitoring</h1><button class="btn btn-primary btn-sm" onclick="loadMetrics()">↻ Refresh</button></div>
    <div class="monitoring-grid">
      <div class="section">
        <div class="section-title">📡 Service Status</div>
        <div class="metric-row"><span class="metric-key">API Gateway</span><span class="metric-val up" id="m-gw">Checking...</span></div>
        <div class="metric-row"><span class="metric-key">user-service</span><span class="metric-val" id="m-us">—</span></div>
        <div class="metric-row"><span class="metric-key">order-service</span><span class="metric-val" id="m-os">—</span></div>
        <div class="metric-row"><span class="metric-key">payment-service</span><span class="metric-val" id="m-ps">—</span></div>
      </div>
      <div class="section">
        <div class="section-title">📊 Gateway Metrics</div>
        <div id="metrics-list"><div style="color:#64748b;font-size:13px">Click Refresh to load metrics</div></div>
      </div>
    </div>
    <div class="section" style="margin-top:16px">
      <div class="section-title">🔗 External Links</div>
      <div style="display:flex;gap:12px;flex-wrap:wrap">
        <a href="http://localhost:9090" target="_blank" class="btn btn-secondary">📈 Prometheus</a>
        <a href="http://localhost:3000" target="_blank" class="btn btn-secondary">📉 Grafana</a>
        <a href="http://localhost:8222" target="_blank" class="btn btn-secondary">📨 NATS Monitor</a>
        <a href="/health" target="_blank" class="btn btn-secondary">❤️ Health</a>
        <a href="/metrics" target="_blank" class="btn btn-secondary">🔢 Raw Metrics</a>
      </div>
    </div>
  </div>
</div>

<script>
const API = '';
let token = localStorage.getItem('token') || '';
let userId = localStorage.getItem('userId') || '';

function updateAuthStatus() {
  const el = document.getElementById('authStatus');
  el.textContent = userId ? 'Logged in (ID: ' + userId + ')' : 'Not logged in';
}
updateAuthStatus();

function showPage(name) {
  document.querySelectorAll('.page-tabs').forEach(p => p.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
  document.getElementById('page-' + name).classList.add('active');
  event.currentTarget.classList.add('active');
  if (name === 'monitoring') loadMetrics();
}

function showAlert(id, msg, type) {
  const el = document.getElementById(id);
  el.textContent = msg;
  el.className = 'alert ' + type + ' show';
  setTimeout(() => el.classList.remove('show'), 5000);
}

async function apiFetch(path, opts = {}) {
  const headers = {'Content-Type': 'application/json', ...(opts.headers || {})};
  if (token) headers['Authorization'] = 'Bearer ' + token;
  const res = await fetch(API + path, {...opts, headers});
  const data = await res.json();
  return {ok: res.ok, status: res.status, data};
}

async function doRegister() {
  const name = document.getElementById('reg-name').value;
  const email = document.getElementById('reg-email').value;
  const password = document.getElementById('reg-password').value;
  const {ok, data} = await apiFetch('/api/auth/register', {method:'POST', body: JSON.stringify({name, email, password})});
  if (ok) {
    token = data.access_token; userId = data.user_id;
    localStorage.setItem('token', token); localStorage.setItem('userId', userId);
    updateAuthStatus();
    showAlert('reg-alert', '✓ Registered successfully! User ID: ' + data.user_id, 'success');
    const tb = document.getElementById('reg-token');
    tb.textContent = 'Access Token: ' + data.access_token;
    tb.classList.add('show');
  } else {
    showAlert('reg-alert', '✗ ' + (data.error || 'Registration failed'), 'error');
  }
}

async function doLogin() {
  const email = document.getElementById('login-email').value;
  const password = document.getElementById('login-password').value;
  const {ok, data} = await apiFetch('/api/auth/login', {method:'POST', body: JSON.stringify({email, password})});
  if (ok) {
    token = data.access_token; userId = data.user_id;
    localStorage.setItem('token', token); localStorage.setItem('userId', userId);
    updateAuthStatus();
    showAlert('login-alert', '✓ Logged in! User ID: ' + data.user_id, 'success');
    const tb = document.getElementById('login-token');
    tb.textContent = 'Access Token: ' + data.access_token;
    tb.classList.add('show');
  } else {
    showAlert('login-alert', '✗ ' + (data.error || 'Login failed'), 'error');
  }
}

async function doLogout() {
  if (token) await apiFetch('/api/auth/logout', {method:'POST'});
  token = ''; userId = '';
  localStorage.removeItem('token'); localStorage.removeItem('userId');
  updateAuthStatus();
}

async function loadUsers() {
  const {ok, data} = await apiFetch('/api/users?page=1&page_size=50');
  const tbody = document.getElementById('users-table');
  if (!ok || !data.users || !data.users.length) {
    tbody.innerHTML = '<tr><td colspan="6" class="empty">No users found</td></tr>'; return;
  }
  tbody.innerHTML = data.users.map(u => '<tr>' +
    '<td>' + (u.id||'') + '</td><td>' + (u.name||'') + '</td><td>' + (u.email||'') + '</td>' +
    '<td>' + (u.is_banned ? '<span class="tag tag-cancelled">Banned</span>' : '<span class="tag tag-active">Active</span>') + '</td>' +
    '<td>' + fmtDate(u.created_at) + '</td>' +
    '<td class="actions"><button class="btn btn-danger btn-sm" onclick="deleteUser(' + u.id + ')">Delete</button></td>' +
    '</tr>').join('');
  document.getElementById('stat-users').textContent = data.total || data.users.length;
}

async function deleteUser(id) {
  if (!confirm('Delete user ' + id + '?')) return;
  const {ok} = await apiFetch('/api/users/' + id, {method:'DELETE'});
  if (ok) { showAlert('users-alert','✓ User deleted','success'); loadUsers(); }
  else showAlert('users-alert','✗ Delete failed','error');
}

async function createOrder() {
  const userid = parseInt(document.getElementById('o-userid').value);
  const product = document.getElementById('o-product').value;
  const qty = parseInt(document.getElementById('o-qty').value) || 1;
  const price = parseFloat(document.getElementById('o-price').value) || 5000;
  const {ok, data} = await apiFetch('/api/orders', {method:'POST', body: JSON.stringify({
    user_id: userid, items: [{product_name: product, quantity: qty, price}]
  })});
  if (ok) { showAlert('order-alert','✓ Order created! ID: ' + data.order?.id,'success'); loadOrders(); }
  else showAlert('order-alert','✗ ' + (data.error||'Failed'),'error');
}

async function loadOrders() {
  const {ok, data} = await apiFetch('/api/orders?page=1&page_size=50');
  const tbody = document.getElementById('orders-table');
  if (!ok || !data.orders || !data.orders.length) {
    tbody.innerHTML = '<tr><td colspan="7" class="empty">No orders found</td></tr>'; return;
  }
  tbody.innerHTML = data.orders.map(o => '<tr>' +
    '<td>' + (o.id||'') + '</td><td>' + (o.user_id||'') + '</td>' +
    '<td><span class="tag tag-' + (o.status||'pending') + '">' + (o.status||'pending') + '</span></td>' +
    '<td>' + fmtMoney(o.total_price) + '</td>' +
    '<td>' + ((o.items||[]).length) + ' items</td>' +
    '<td>' + fmtDate(o.created_at) + '</td>' +
    '<td class="actions"><button class="btn btn-danger btn-sm" onclick="cancelOrder(' + o.id + ')">Cancel</button></td>' +
    '</tr>').join('');
  document.getElementById('stat-orders').textContent = data.total || data.orders.length;
}

async function cancelOrder(id) {
  const {ok} = await apiFetch('/api/orders/' + id + '/cancel', {method:'POST'});
  if (ok) { showAlert('order-alert','✓ Order cancelled','success'); loadOrders(); }
  else { const {ok:ok2} = await apiFetch('/api/orders/' + id, {method:'DELETE'}); loadOrders(); }
}

async function loadPayments() {
  const {ok, data} = await apiFetch('/api/payments?page=1&page_size=50');
  const tbody = document.getElementById('payments-table');
  if (!ok || !data.payments || !data.payments.length) {
    tbody.innerHTML = '<tr><td colspan="8" class="empty">No payments found</td></tr>'; return;
  }
  tbody.innerHTML = data.payments.map(p => '<tr>' +
    '<td>' + (p.id||'') + '</td><td>' + (p.order_id||'') + '</td><td>' + (p.user_id||'') + '</td>' +
    '<td>' + fmtMoney(p.amount) + '</td><td>' + (p.currency||'KZT') + '</td>' +
    '<td><span class="tag tag-' + (p.status||'pending') + '">' + (p.status||'pending') + '</span></td>' +
    '<td>' + (p.payment_method||'card') + '</td>' +
    '<td>' + fmtDate(p.created_at) + '</td>' +
    '</tr>').join('');
  document.getElementById('stat-payments').textContent = data.total || data.payments.length;
}

async function loadMetrics() {
  // Check gateway health
  try {
    const r = await fetch('/health');
    const d = await r.json();
    document.getElementById('m-gw').textContent = d.status === 'ok' ? '✓ UP' : '✗ DOWN';
    document.getElementById('m-gw').className = 'metric-val up';
  } catch { document.getElementById('m-gw').textContent = '✗ DOWN'; }

  // Check services via users/orders/payments API
  for (const [id, path] of [['m-us','/api/users?page=1&page_size=1'],['m-os','/api/orders?page=1&page_size=1'],['m-ps','/api/payments?page=1&page_size=1']]) {
    try {
      const r = await fetch(API + path, {headers: token ? {Authorization:'Bearer '+token} : {}});
      document.getElementById(id).textContent = r.ok ? '✓ UP' : '✗ DOWN';
      document.getElementById(id).className = r.ok ? 'metric-val up' : 'metric-val';
    } catch { document.getElementById(id).textContent = '✗ DOWN'; }
  }

  // Raw metrics
  try {
    const r = await fetch('/metrics');
    const txt = await r.text();
    const lines = txt.split('\n').filter(l => l && !l.startsWith('#'));
    document.getElementById('metrics-list').innerHTML = lines.map(l => {
      const [k,v] = l.split(' ');
      return '<div class="metric-row"><span class="metric-key">' + k + '</span><span class="metric-val">' + v + '</span></div>';
    }).join('') || '<div style="color:#64748b;font-size:13px">No metrics</div>';
  } catch {}
}

function fmtDate(s) {
  if (!s) return '—';
  try { return new Date(s).toLocaleString('ru-RU', {day:'2-digit',month:'2-digit',year:'2-digit',hour:'2-digit',minute:'2-digit'}); }
  catch { return s; }
}

function fmtMoney(v) {
  if (v == null) return '—';
  return parseFloat(v).toLocaleString('ru-RU') + ' ₸';
}

// Load dashboard stats on startup
(async () => {
  try { const {data} = await apiFetch('/api/users?page=1&page_size=1'); if (data.total) document.getElementById('stat-users').textContent = data.total; } catch {}
  try { const {data} = await apiFetch('/api/orders?page=1&page_size=1'); if (data.total) document.getElementById('stat-orders').textContent = data.total; } catch {}
  try { const {data} = await apiFetch('/api/payments?page=1&page_size=1'); if (data.total) document.getElementById('stat-payments').textContent = data.total; } catch {}
})();
</script>
</body>
</html>`
