let currentSelectedSlug = null;
let autoRefreshInterval = null;

document.addEventListener('DOMContentLoaded', () => {
    loadLinks();

    const createForm = document.getElementById('createLinkForm');
    createForm.addEventListener('submit', handleCreateLink);

    const refreshBtn = document.getElementById('refreshBtn');
    refreshBtn.addEventListener('click', () => {
        loadLinks();
        if (currentSelectedSlug) {
            loadAnalytics(currentSelectedSlug);
        }
    });

    const copyBtn = document.getElementById('copyTrackLinkBtn');
    copyBtn.addEventListener('click', handleCopyLink);

    // Auto-refresh analytics if a link is selected every 5s
    autoRefreshInterval = setInterval(() => {
        if (currentSelectedSlug) {
            loadAnalytics(currentSelectedSlug, false);
            loadLinks(false);
        }
    }, 5000);
});

async function handleCreateLink(e) {
    e.preventDefault();
    const targetUrlInput = document.getElementById('targetUrl');
    const customSlugInput = document.getElementById('customSlug');
    const alertBox = document.getElementById('createAlert');
    const createBtn = document.getElementById('createBtn');

    alertBox.className = 'alert hidden';
    createBtn.disabled = true;

    const payload = {
        target_url: targetUrlInput.value.trim(),
        custom_slug: customSlugInput.value.trim()
    };

    try {
        const res = await fetch('/v1/links', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        const data = await res.json();

        if (!res.ok) {
            let errorMsg = 'Failed to create trackable link';
            if (data.error) {
                if (typeof data.error === 'object') {
                    errorMsg = Object.values(data.error).join(', ');
                } else {
                    errorMsg = data.error;
                }
            }
            throw new Error(errorMsg);
        }

        const trackUrl = data.link.short_url || `${window.location.origin}/r/${data.link.slug}`;
        alertBox.textContent = `Success! Link created: ${trackUrl}`;
        alertBox.className = 'alert alert-success';

        // Clear input
        targetUrlInput.value = '';
        customSlugInput.value = '';

        await loadLinks();
        selectLink(data.link.slug);
    } catch (err) {
        alertBox.textContent = err.message;
        alertBox.className = 'alert alert-error';
    } finally {
        createBtn.disabled = false;
    }
}

async function loadLinks(updateSelectionUI = true) {
    try {
        const res = await fetch('/v1/links');
        if (!res.ok) return;

        const data = await res.json();
        const links = data.links || [];
        const linksListEl = document.getElementById('linksList');

        const countBadge = document.getElementById('linksCountBadge');
        if (countBadge) {
            countBadge.textContent = links.length;
        }

        if (links.length === 0) {
            linksListEl.innerHTML = `
                <div class="empty-state">
                    <p class="empty-title">No endpoints configured</p>
                    <p class="empty-desc">Create your first tracking link above.</p>
                </div>
            `;
            return;
        }

        linksListEl.innerHTML = '';
        links.forEach(link => {
            const item = document.createElement('div');
            item.className = `link-item ${currentSelectedSlug === link.slug ? 'active' : ''}`;
            item.onclick = () => selectLink(link.slug);

            item.innerHTML = `
                <div class="link-item-header">
                    <span class="link-item-slug">/r/${escapeHtml(link.slug)}</span>
                    <span class="link-item-visits">${link.total_visits} ${link.total_visits === 1 ? 'visit' : 'visits'}</span>
                </div>
                <div class="link-item-target" title="${escapeHtml(link.target_url)}">
                    ${escapeHtml(link.target_url)}
                </div>
            `;
            linksListEl.appendChild(item);
        });

        if (!currentSelectedSlug && links.length > 0 && updateSelectionUI) {
            selectLink(links[0].slug);
        }
    } catch (err) {
        console.error('Failed to load links:', err);
    }
}

async function selectLink(slug) {
    currentSelectedSlug = slug;

    // Update active highlight in list
    const items = document.querySelectorAll('.link-item');
    items.forEach(el => {
        const text = el.querySelector('.link-item-slug')?.textContent;
        if (text === `/r/${slug}`) {
            el.classList.add('active');
        } else {
            el.classList.remove('active');
        }
    });

    await loadAnalytics(slug);
}

async function loadAnalytics(slug, showLoading = true) {
    try {
        const res = await fetch(`/v1/links/${slug}`);
        if (!res.ok) return;

        const data = await res.json();
        const link = data.link;
        if (!link) return;

        const noSelection = document.getElementById('noSelectionState');
        const analyticsContent = document.getElementById('analyticsContent');

        noSelection.classList.add('hidden');
        analyticsContent.classList.remove('hidden');

        document.getElementById('selectedSlugBadge').textContent = `/r/${link.slug}`;
        document.getElementById('selectedTargetUrl').textContent = link.target_url;
        document.getElementById('totalVisitsCount').textContent = link.total_visits;

        const fullTrackUrl = link.short_url || `${window.location.origin}/r/${link.slug}`;
        const testBtn = document.getElementById('testTrackLinkBtn');
        testBtn.href = fullTrackUrl;

        // Calculate metrics and distributions
        const ips = new Set();
        const countries = new Set();
        const osCounts = {};
        const browserCounts = {};
        const deviceCounts = {};
        const countryCounts = {};

        const visits = link.visits || [];
        visits.forEach(v => {
            if (v.ip) ips.add(v.ip);
            if (v.country && v.country !== 'Unknown') {
                countries.add(v.country);
                countryCounts[v.country] = (countryCounts[v.country] || 0) + 1;
            }
            if (v.os) osCounts[v.os] = (osCounts[v.os] || 0) + 1;
            if (v.browser) browserCounts[v.browser] = (browserCounts[v.browser] || 0) + 1;
            if (v.device_type) deviceCounts[v.device_type] = (deviceCounts[v.device_type] || 0) + 1;
        });

        document.getElementById('uniqueIPsCount').textContent = ips.size;
        document.getElementById('uniqueCountriesCount').textContent = countries.size;

        // Top OS
        const sortedOS = Object.entries(osCounts).sort((a, b) => b[1] - a[1]);
        document.getElementById('topOSStat').textContent = sortedOS.length > 0 ? sortedOS[0][0] : '-';

        // Render breakdown cards
        renderBreakdown('osBreakdownList', osCounts);
        renderBreakdown('browserBreakdownList', browserCounts);
        renderBreakdown('deviceBreakdownList', deviceCounts);
        renderBreakdown('countryBreakdownList', countryCounts);

        // Render table
        const tbody = document.getElementById('visitsTableBody');
        if (visits.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" class="table-empty">No visits recorded for this endpoint yet.</td></tr>';
        } else {
            tbody.innerHTML = '';
            visits.forEach(v => {
                const tr = document.createElement('tr');
                
                const timeStr = new Date(v.visited_at).toISOString().replace('T', ' ').substring(0, 19);
                let locStr = 'Unknown';
                if (v.city && v.country && v.city !== 'Unknown') {
                    locStr = `${escapeHtml(v.city)}, ${escapeHtml(v.country)}`;
                } else if (v.country && v.country !== 'Unknown') {
                    locStr = escapeHtml(v.country);
                }

                const ispStr = v.isp || v.org || 'Unknown';
                const osStr = v.os || 'Unknown';
                const browserStr = v.browser || 'Unknown';
                const deviceStr = v.device_type || 'Desktop';
                const subMeta = v.referer ? `Ref: ${escapeHtml(v.referer)}` : escapeHtml(v.user_agent || 'Unknown');

                tr.innerHTML = `
                    <td class="time-col">${escapeHtml(timeStr)}</td>
                    <td><span class="ip-tag">${escapeHtml(v.ip || 'N/A')}</span></td>
                    <td class="location-col">${locStr}</td>
                    <td><span class="isp-text" title="${escapeHtml(ispStr)}">${escapeHtml(ispStr)}</span></td>
                    <td>
                        <span class="os-tag">${escapeHtml(osStr)}</span>
                        <span class="browser-tag">${escapeHtml(browserStr)}</span>
                    </td>
                    <td><span class="device-badge">${escapeHtml(deviceStr)}</span></td>
                    <td><div class="ua-col" title="${escapeHtml(v.user_agent)}">${subMeta}</div></td>
                `;
                tbody.appendChild(tr);
            });
        }
    } catch (err) {
        console.error('Failed to load analytics:', err);
    }
}

function renderBreakdown(elementId, countsMap) {
    const el = document.getElementById(elementId);
    if (!el) return;

    const entries = Object.entries(countsMap).sort((a, b) => b[1] - a[1]);
    if (entries.length === 0) {
        el.innerHTML = '<span class="text-dim">No data</span>';
        return;
    }

    el.innerHTML = '';
    entries.slice(0, 5).forEach(([name, count]) => {
        const row = document.createElement('div');
        row.className = 'breakdown-row';
        row.innerHTML = `
            <span class="breakdown-name" title="${escapeHtml(name)}">${escapeHtml(name)}</span>
            <span class="breakdown-count">${count}</span>
        `;
        el.appendChild(row);
    });
}

function handleCopyLink() {
    if (!currentSelectedSlug) return;
    const testBtn = document.getElementById('testTrackLinkBtn');
    const fullTrackUrl = testBtn.href || `${window.location.origin}/r/${currentSelectedSlug}`;
    navigator.clipboard.writeText(fullTrackUrl).then(() => {
        const copyBtn = document.getElementById('copyTrackLinkBtn');
        const origText = copyBtn.textContent;
        copyBtn.textContent = 'Copied!';
        setTimeout(() => {
            copyBtn.textContent = origText;
        }, 1500);
    });
}

function escapeHtml(str) {
    if (!str) return '';
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}
