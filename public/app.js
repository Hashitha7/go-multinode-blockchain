const API = '';
let currentTab = 'dashboard';
let refreshTimer = null;

// ── DOM refs ──
const refreshBtn = document.getElementById('refresh-btn');
const lastUpdateText = document.getElementById('last-update-text');
const pageTitle = document.getElementById('page-title');
const topbarAddr = document.getElementById('topbar-addr');
const sidebarAddr = document.getElementById('sidebar-addr');

// Stat refs
const sHeight = document.getElementById('s-height');
const sPeers = document.getElementById('s-peers');
const sMempool = document.getElementById('s-mempool');
const sHeadHash = document.getElementById('s-head-hash');
const sChainLen = document.getElementById('s-chain-len');

// ── Tab Navigation ──
document.querySelectorAll('.nav-item').forEach(item => {
    item.addEventListener('click', e => {
        e.preventDefault();
        document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
        document.querySelectorAll('.tab-page').forEach(p => p.classList.remove('active'));
        item.classList.add('active');
        currentTab = item.dataset.tab;
        const page = document.getElementById(`page-${currentTab}`);
        if (page) page.classList.add('active');
        pageTitle.textContent = item.textContent.trim();
        fetchAll();
    });
});

// ── Helpers ──
function shortHash(h) {
    if (!h || h.length < 16) return h || '—';
    return h.substring(0, 14) + '...' + h.substring(h.length - 10);
}

function formatTime() {
    return new Date().toLocaleTimeString('en-US', { hour12: false });
}

// ── Fetches ──
async function fetchStatus() {
    const res = await fetch(`${API}/status`);
    if (!res.ok) throw new Error('offline');
    return res.json();
}

async function fetchChain() {
    const res = await fetch(`${API}/chain`);
    if (!res.ok) throw new Error('chain failed');
    return res.json();
}

async function fetchPeers() {
    const res = await fetch(`${API}/peers`);
    if (!res.ok) throw new Error('peers failed');
    return res.json();
}

async function fetchBalances() {
    const res = await fetch(`${API}/balances`);
    if (!res.ok) throw new Error('balances failed');
    return res.json();
}

async function fetchMempool() {
    const res = await fetch(`${API}/mempool`);
    if (!res.ok) throw new Error('mempool failed');
    return res.json();
}

// ── Renderers ──
function renderDashboard(status) {
    sHeight.textContent = status.height ?? '—';
    sPeers.textContent = status.peer_count ?? '—';
    sMempool.textContent = status.mempool_size ?? '—';
    sHeadHash.textContent = status.head_hash ?? '—';
    sChainLen.textContent = `Chain length: ${status.chain_length ?? '—'}`;
    const addr = status.address ?? '—';
    topbarAddr.textContent = addr;
    sidebarAddr.textContent = addr.replace('localhost', '');
}

function renderBlocks(blocks) {
    const tbody = document.getElementById('blocks-tbody');
    const pill = document.getElementById('blocks-pill');
    const list = (blocks || []).slice().reverse().slice(0, 30);
    pill.textContent = `${blocks.length} blocks`;
    if (list.length === 0) {
        tbody.innerHTML = `<tr><td colspan="4"><div class="empty-state"><i class="ph ph-stack"></i><p>No blocks yet</p></div></td></tr>`;
        return;
    }
    tbody.innerHTML = list.map(b => `
        <tr>
            <td class="index-cell"><span>#${b.index}</span></td>
            <td class="hash-cell" title="${b.hash}">${shortHash(b.hash)}</td>
            <td class="miner-cell"><span class="miner-badge"><i class="ph ph-hammer"></i>${shortHash(b.miner)}</span></td>
            <td class="tx-count-cell"><span class="tx-badge">${(b.transactions || []).length} txns</span></td>
        </tr>`).join('');
}

function renderPeers(data) {
    const grid = document.getElementById('peers-grid');
    const pill = document.getElementById('peers-pill');
    const peers = data.peers || [];
    pill.textContent = `${peers.length} peers`;
    if (peers.length === 0) {
        grid.innerHTML = `<div class="empty-state" style="grid-column:1/-1"><i class="ph ph-wifi-slash"></i><p>No peers connected</p></div>`;
        return;
    }
    grid.innerHTML = peers.map(p => `
        <div class="peer-card">
            <div class="peer-dot"></div>
            <span class="peer-addr">${p}</span>
        </div>`).join('');
}

function renderBalances(data) {
    const tbody = document.getElementById('balances-tbody');
    const pill = document.getElementById('balances-pill');
    const entries = Object.entries(data || {}).sort((a,b) => b[1]-a[1]);
    pill.textContent = `${entries.length} accounts`;
    if (entries.length === 0) {
        tbody.innerHTML = `<tr><td colspan="2"><div class="empty-state"><i class="ph ph-coin"></i><p>No balances found</p></div></td></tr>`;
        return;
    }
    tbody.innerHTML = entries.map(([addr, bal]) => `
        <tr>
            <td class="hash-cell" title="${addr}">${shortHash(addr)}</td>
            <td class="amount-cell">${bal.toLocaleString()}</td>
        </tr>`).join('');
}

function renderMempool(data) {
    const tbody = document.getElementById('mempool-tbody');
    const pill = document.getElementById('mempool-pill');
    const txs = data.transactions || [];
    pill.textContent = `${txs.length} txns`;
    if (txs.length === 0) {
        tbody.innerHTML = `<tr><td colspan="5"><div class="empty-state"><i class="ph ph-check-circle" style="color:#10b981;opacity:1"></i><p style="color:#10b981">Mempool empty — all transactions confirmed!</p></div></td></tr>`;
        return;
    }
    tbody.innerHTML = txs.map(tx => `
        <tr>
            <td class="hash-cell" title="${tx.id}">${shortHash(tx.id)}</td>
            <td class="hash-cell" title="${tx.from}">${shortHash(tx.from)}</td>
            <td class="hash-cell" title="${tx.to}">${shortHash(tx.to)}</td>
            <td class="amount-cell">${tx.amount}</td>
            <td style="color:#6b7a99">${tx.fee}</td>
        </tr>`).join('');
}

// ── Master Fetch ──
async function fetchAll() {
    refreshBtn.classList.add('spinning');
    const dot = document.querySelector('.node-status-badge .dot');
    const statusText = document.getElementById('sidebar-status');

    try {
        const status = await fetchStatus();
        renderDashboard(status);
        dot.style.background = 'var(--green)';
        statusText.textContent = 'Online';
    } catch(e) {
        dot.style.background = 'var(--red)';
        statusText.textContent = 'Offline';
    }

    if (currentTab === 'blocks') {
        try { const d = await fetchChain(); renderBlocks(d); } catch(e) {}
    } else if (currentTab === 'peers') {
        try { const d = await fetchPeers(); renderPeers(d); } catch(e) {}
    } else if (currentTab === 'balances') {
        try { const d = await fetchBalances(); renderBalances(d); } catch(e) {}
    } else if (currentTab === 'mempool') {
        try { const d = await fetchMempool(); renderMempool(d); } catch(e) {}
    }

    lastUpdateText.textContent = formatTime();
    refreshBtn.classList.remove('spinning');
}

refreshBtn.addEventListener('click', fetchAll);

// Init + auto-refresh every 5s
fetchAll();
setInterval(fetchAll, 5000);
