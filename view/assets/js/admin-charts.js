let state = {
  metric: "items",
  days: 30,
  mode: "count",
  granularity: "daily",
  points: [],
  metricStats: null,
  userStats: null,
  loading: false,
  fallback: false,
};

const METRIC_META = {
  items: { labelKey: "admin.chartsMetricItems", fallback: "Total items", color: "#e50914" },
  users: { labelKey: "admin.chartsMetricUsers", fallback: "Users", color: "#2563eb" },
  communities: { labelKey: "admin.chartsMetricCommunities", fallback: "Communities", color: "#0891b2" },
  categories: { labelKey: "admin.chartsMetricCategories", fallback: "Categories", color: "#16a34a" },
  movies: { labelKey: "admin.chartsMetricMovies", fallback: "Movies", color: "#e50914" },
  tv_shows: { labelKey: "admin.chartsMetricTvShows", fallback: "TV Shows", color: "#7c3aed" },
  games: { labelKey: "admin.chartsMetricGames", fallback: "Games", color: "#16a34a" },
  books: { labelKey: "admin.chartsMetricBooks", fallback: "Books", color: "#f5c518" },
  subscriptions: { labelKey: "admin.chartsMetricSubscriptions", fallback: "Subscriptions", color: "#0ea5e9" },
  subscription_revenue: { labelKey: "admin.chartsMetricSubscriptionRevenue", fallback: "Subscription revenue", color: "#22c55e" },
};

function isMoneyMetric() {
  return state.metric === "subscription_revenue";
}

function getMetricLabel(metric) {
  const meta = METRIC_META[metric] || METRIC_META.items;
  const label = t(meta.labelKey);
  return label && label !== meta.labelKey ? label : meta.fallback;
}

function getQueryParam(name, fallback) {
  const params = new URLSearchParams(window.location.search);
  return params.get(name) || fallback;
}

function setQueryParam(name, value) {
  const params = new URLSearchParams(window.location.search);
  params.set(name, value);
  const newUrl = `${window.location.pathname}?${params.toString()}`;
  window.history.replaceState({}, "", newUrl);
}

function valueOf(point) {
  return Number(point.count) || 0;
}

function parsePeriod(period) {
  const parts = String(period || "").split("-").map(Number);
  if (parts.length >= 2 && state.granularity === "monthly") {
    return new Date(parts[0], parts[1] - 1, 1);
  }
  if (parts.length >= 3) {
    return new Date(parts[0], parts[1] - 1, parts[2]);
  }
  return new Date(period);
}

function formatPeriod(period) {
  const date = parsePeriod(period);
  if (state.granularity === "monthly") {
    return date.toLocaleDateString(undefined, { month: "short", year: "2-digit" });
  }
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function formatNumber(value) {
  if (isMoneyMetric()) {
    if (value >= 1000000) return `${(value / 1000000).toFixed(value >= 10000000 ? 0 : 1)}m`;
    if (value >= 1000) return `${(value / 1000).toFixed(value >= 10000 ? 0 : 1)}k`;
    return String(Math.round(value));
  }
  if (value >= 1000) return `${(value / 1000).toFixed(value >= 10000 ? 0 : 1)}k`;
  return String(value);
}

function formatFull(value) {
  return new Intl.NumberFormat().format(value);
}

function formatMoney(value, currency) {
  const amount = Number(value) || 0;
  const code = currency || "USD";
  try {
    return new Intl.NumberFormat(undefined, { style: "currency", currency: code }).format(amount);
  } catch {
    return `${formatFull(amount)} ${code}`;
  }
}

function renderHeading() {
  const heading = document.getElementById("chartsHeading");
  const panelTitle = document.getElementById("chartPanelTitle");
  const label = getMetricLabel(state.metric);
  if (heading) heading.textContent = label;
  if (panelTitle) {
    panelTitle.textContent = `${label} · ${isMoneyMetric() ? t("admin.chartsModeRevenue") : t("admin.chartsModeCount")}`;
  }
}

function renderControls() {
  const subscriptionSwitcher = document.getElementById("subscriptionMetricSwitcher");
  const showSubscriptionSwitcher = state.metric === "subscriptions" || state.metric === "subscription_revenue";
  if (subscriptionSwitcher) subscriptionSwitcher.hidden = !showSubscriptionSwitcher;

  document.querySelectorAll("#subscriptionMetricSwitcher .segmented-btn").forEach((btn) => {
    const active = btn.dataset.subscriptionMetric === state.metric;
    btn.classList.toggle("active", active);
    btn.setAttribute("aria-pressed", String(active));
  });

  document.querySelectorAll("#periodSwitcher .segmented-btn").forEach((btn) => {
    const active = Number(btn.dataset.days) === state.days;
    btn.classList.toggle("active", active);
    btn.setAttribute("aria-pressed", String(active));
  });
}

function niceMax(value) {
  if (value <= 0) return 1;
  const magnitude = Math.pow(10, Math.floor(Math.log10(value)));
  const normalized = value / magnitude;
  let step;
  if (normalized <= 1) step = 1;
  else if (normalized <= 2) step = 2;
  else if (normalized <= 5) step = 5;
  else step = 10;
  return step * magnitude;
}

function renderLineChart() {
  const wrap = document.getElementById("lineChart");
  if (!wrap) return;
  wrap.innerHTML = "";

  if (state.loading) {
    wrap.innerHTML = `
      <div class="chart-loading" aria-live="polite">
        <span></span><span></span><span></span>
        <p>${escapeHtml(t("admin.chartsLoading"))}</p>
      </div>
    `;
    return;
  }

  const points = state.points;
  if (!points.length) {
    wrap.appendChild(emptyState("admin.chartsNoData"));
    return;
  }

  const values = points.map(valueOf);
  const rawMax = Math.max(...values, 0);
  const yMax = niceMax(Math.max(rawMax * 1.12, 1));

  const width = 880;
  const height = 340;
  const padding = { top: 24, right: 24, bottom: 38, left: 52 };
  const innerW = width - padding.left - padding.right;
  const innerH = height - padding.top - padding.bottom;

  const xStep = points.length > 1 ? innerW / (points.length - 1) : 0;
  const xFor = (index) => padding.left + (points.length > 1 ? index * xStep : innerW / 2);
  const yFor = (value) => padding.top + innerH - (value / yMax) * innerH;

  const color = METRIC_META[state.metric]?.color || "#e50914";
  const gradientId = "lineFill";

  // Y gridlines + labels (4 ticks)
  const ticks = 4;
  let grid = "";
  for (let i = 0; i <= ticks; i++) {
    const value = (yMax / ticks) * i;
    const y = yFor(value);
    grid += `<line class="lc-grid" x1="${padding.left}" y1="${y}" x2="${width - padding.right}" y2="${y}" />`;
    grid += `<text class="lc-axis-label" x="${padding.left - 10}" y="${y + 4}" text-anchor="end">${formatNumber(Math.round(value))}</text>`;
  }

  // X labels (up to ~6 evenly spaced)
  const labelCount = Math.min(6, points.length);
  let xLabels = "";
  for (let i = 0; i < labelCount; i++) {
    const index = labelCount === 1 ? 0 : Math.round((i / (labelCount - 1)) * (points.length - 1));
    xLabels += `<text class="lc-axis-label" x="${xFor(index)}" y="${height - padding.bottom + 22}" text-anchor="middle">${escapeHtml(formatPeriod(points[index].period))}</text>`;
  }

  // Line + area paths
  const coords = points.map((point, index) => ({ x: xFor(index), y: yFor(valueOf(point)) }));
  const smoothCommands = coords.slice(1).map((point, index) => {
    const prev = coords[index];
    const midX = (prev.x + point.x) / 2;
    return `C ${midX},${prev.y} ${midX},${point.y} ${point.x},${point.y}`;
  }).join(" ");
  const linePath = coords.length === 1 ? `M ${coords[0].x},${coords[0].y}` : `M ${coords[0].x},${coords[0].y} ${smoothCommands}`;
  const areaPath = coords.length === 1
    ? `M ${coords[0].x},${yFor(0)} L ${coords[0].x},${coords[0].y} L ${coords[0].x},${yFor(0)} Z`
    : `M ${coords[0].x},${yFor(0)} L ${coords[0].x},${coords[0].y} ${smoothCommands} L ${coords[coords.length - 1].x},${yFor(0)} Z`;

  const dots = points
    .map((point, index) => {
      const cx = xFor(index);
      const cy = yFor(valueOf(point));
      return `<circle class="lc-point" cx="${cx}" cy="${cy}" r="3" stroke="${color}" data-index="${index}" />`;
    })
    .join("");

  const svg = `
    <svg class="line-chart" viewBox="0 0 ${width} ${height}" preserveAspectRatio="xMidYMid meet">
      <defs>
        <linearGradient id="${gradientId}" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="${color}" stop-opacity="0.32" />
          <stop offset="100%" stop-color="${color}" stop-opacity="0" />
        </linearGradient>
      </defs>
      ${grid}
      ${xLabels}
      <path class="lc-area" d="${areaPath}" fill="url(#${gradientId})" />
      <path class="lc-line" d="${linePath}" fill="none" stroke="${color}" />
      ${dots}
      <line class="lc-hover-line" x1="0" y1="${padding.top}" x2="0" y2="${height - padding.bottom}" />
      <circle class="lc-hover-point" r="5" />
    </svg>
    <div class="lc-tooltip" hidden></div>
  `;

  wrap.innerHTML = svg;
  bindChartHover(wrap, points, xFor, yFor, padding, width, height);
}

function bindChartHover(wrap, points, xFor, yFor, padding, width, height) {
  const svg = wrap.querySelector(".line-chart");
  const tooltip = wrap.querySelector(".lc-tooltip");
  const hoverLine = wrap.querySelector(".lc-hover-line");
  const hoverPoint = wrap.querySelector(".lc-hover-point");
  if (!svg || !tooltip || !hoverLine || !hoverPoint) return;

  function move(event) {
    const rect = svg.getBoundingClientRect();
    const scaleX = width / rect.width;
    const pointerX = (event.clientX - rect.left) * scaleX;

    let nearest = 0;
    let nearestDist = Infinity;
    for (let i = 0; i < points.length; i++) {
      const dist = Math.abs(xFor(i) - pointerX);
      if (dist < nearestDist) {
        nearestDist = dist;
        nearest = i;
      }
    }

    const cx = xFor(nearest);
    const cy = yFor(valueOf(points[nearest]));
    hoverLine.setAttribute("x1", cx);
    hoverLine.setAttribute("x2", cx);
    hoverPoint.setAttribute("cx", cx);
    hoverPoint.setAttribute("cy", cy);
    hoverPoint.setAttribute("stroke", METRIC_META[state.metric]?.color || "#e50914");

    const value = valueOf(points[nearest]);
    const label = getMetricLabel(state.metric);
    tooltip.hidden = false;
    tooltip.innerHTML = `
      <strong>${escapeHtml(isMoneyMetric() ? formatMoney(value, state.metricStats?.currency) : formatFull(value))}</strong>
      <span>${escapeHtml(label)}</span>
      <em>${escapeHtml(formatPeriod(points[nearest].period))}</em>
    `;

    const tipLeft = (cx / width) * rect.width;
    let offset = 14;
    if (tipLeft > rect.width / 2) offset = -14 - tooltip.offsetWidth;
    tooltip.style.left = `${tipLeft + offset}px`;
    tooltip.style.top = `${(cy / height) * rect.height - tooltip.offsetHeight / 2}px`;
  }

  function hide() {
    tooltip.hidden = true;
  }

  svg.addEventListener("mousemove", move);
  svg.addEventListener("mouseleave", hide);
}

function renderSummary() {
  const container = document.getElementById("chartSummary");
  if (!container) return;

  if (state.loading) {
    container.innerHTML = `
      <div class="summary-card summary-card-loading"></div>
    `;
    return;
  }

  const points = state.points;
  const countSum = points.reduce((sum, point) => sum + (Number(point.count) || 0), 0);

  let cards = [
    { labelKey: "admin.chartsSumNew", value: formatFull(countSum) },
  ];

  if (state.metric === "users" && state.userStats) {
    const registration = state.userStats.registration_stats || {};
    const subscription = state.userStats.subscription_stats || {};
    cards = [
      { labelKey: "admin.chartsUsersToday", value: formatFull(Number(registration.today) || 0) },
      { labelKey: "admin.chartsUsersMonth", value: formatFull(Number(registration.month) || 0) },
      { labelKey: "admin.chartsUsersYear", value: formatFull(Number(registration.year) || 0) },
      { labelKey: "admin.chartsUsersTotal", value: formatFull(Number(registration.total ?? state.userStats.total_users) || 0) },
      { labelKey: "admin.chartsSubscriptionPrice", value: formatMoney(subscription.plan_price, subscription.currency) },
      { labelKey: "admin.chartsSubscriptionsToday", value: formatFull(Number(subscription.today) || 0) },
      { labelKey: "admin.chartsSubscriptionsMonth", value: formatFull(Number(subscription.month) || 0) },
      { labelKey: "admin.chartsSubscriptionsYear", value: formatFull(Number(subscription.year) || 0) },
      { labelKey: "admin.chartsSubscriptionsTotal", value: formatFull(Number(subscription.total) || 0) },
      { labelKey: "admin.chartsSubscriptionRevenueToday", value: formatMoney(subscription.revenue_today, subscription.currency) },
      { labelKey: "admin.chartsSubscriptionRevenueMonth", value: formatMoney(subscription.revenue_month, subscription.currency) },
      { labelKey: "admin.chartsSubscriptionRevenueYear", value: formatMoney(subscription.revenue_year, subscription.currency) },
      { labelKey: "admin.chartsSubscriptionRevenueTotal", value: formatMoney(subscription.revenue_total, subscription.currency) },
    ];
  } else if (state.metricStats) {
    if (isMoneyMetric()) {
      cards = [
        { labelKey: "admin.chartsSubscriptionRevenueToday", value: formatMoney(state.metricStats.today, state.metricStats.currency) },
        { labelKey: "admin.chartsSubscriptionRevenueMonth", value: formatMoney(state.metricStats.month, state.metricStats.currency) },
        { labelKey: "admin.chartsSubscriptionRevenueYear", value: formatMoney(state.metricStats.year, state.metricStats.currency) },
        { labelKey: "admin.chartsSubscriptionRevenueTotal", value: formatMoney(state.metricStats.total, state.metricStats.currency) },
      ];
    } else if (state.metric === "subscriptions") {
      cards = [
        { labelKey: "admin.chartsSubscriptionsToday", value: formatFull(Number(state.metricStats.today) || 0) },
        { labelKey: "admin.chartsSubscriptionsMonth", value: formatFull(Number(state.metricStats.month) || 0) },
        { labelKey: "admin.chartsSubscriptionsYear", value: formatFull(Number(state.metricStats.year) || 0) },
        { labelKey: "admin.chartsSubscriptionsTotal", value: formatFull(Number(state.metricStats.total) || 0) },
      ];
    } else {
      cards = [
        { labelKey: "admin.chartsMetricToday", value: formatFull(Number(state.metricStats.today) || 0) },
        { labelKey: "admin.chartsMetricMonth", value: formatFull(Number(state.metricStats.month) || 0) },
        { labelKey: "admin.chartsMetricYear", value: formatFull(Number(state.metricStats.year) || 0) },
        { labelKey: "admin.chartsMetricTotal", value: formatFull(Number(state.metricStats.total) || 0) },
      ];
    }
  }

  container.innerHTML = cards
    .map(
      (card) => `
        <div class="summary-card">
          <span class="summary-label">${escapeHtml(t(card.labelKey))}</span>
          <strong class="summary-value">${escapeHtml(card.value)}</strong>
        </div>
      `,
    )
    .join("");
}

function emptyState(key) {
  const wrap = document.createElement("p");
  wrap.className = "notice";
  wrap.textContent = t(key);
  return wrap;
}

function renderAll() {
  renderHeading();
  renderControls();
  renderLineChart();
  renderSummary();
}

function currentMetricTotal(stats) {
  if (state.metric === "items") return Number(stats?.total_items) || 0;
  if (state.metric === "users") return Number(stats?.total_users) || 0;
  if (state.metric === "communities") return Number(stats?.total_communities) || 0;
  if (state.metric === "categories") return Number(stats?.total_categories) || 0;
  if (state.metric === "subscriptions") return Number(stats?.user_stats?.subscription_stats?.total) || 0;
  if (state.metric === "subscription_revenue") return Number(stats?.user_stats?.subscription_stats?.revenue_total) || 0;

  const typeByMetric = {
    movies: "movie",
    tv_shows: "tv_show",
    games: "game",
    books: "book",
  };
  const type = typeByMetric[state.metric];
  const entry = (stats?.type_counts || []).find((item) => item.type === type);
  return Number(entry?.count) || 0;
}

function syntheticPeriods(total) {
  const points = [];
  const now = new Date();
  const monthly = state.days > 90;
  state.granularity = monthly ? "monthly" : "daily";

  if (monthly) {
    const months = Math.max(1, Math.floor(state.days / 30));
    const start = new Date(now.getFullYear(), now.getMonth() - months, 1);
    for (let date = start; date <= now; date = new Date(date.getFullYear(), date.getMonth() + 1, 1)) {
      points.push({
        period: `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}`,
        count: 0,
        total,
      });
    }
    return points;
  }

  for (let i = state.days - 1; i >= 0; i--) {
    const date = new Date(now);
    date.setDate(now.getDate() - i);
    points.push({
      period: `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`,
      count: 0,
      total,
    });
  }
  return points;
}

async function loadFallbackChartData(originalError) {
  try {
    const data = await fetchJson("/admin/stats", { headers: authHeaders() });
    const total = currentMetricTotal(data?.stats || {});
    state.metricStats = { today: 0, month: 0, year: 0, total, currency: data?.stats?.user_stats?.subscription_stats?.currency };
    state.userStats = state.metric === "users" ? data?.stats?.user_stats || null : null;
    state.points = syntheticPeriods(total);
    state.loading = false;
    state.fallback = true;
    renderAll();
    setNotice("pageError", t("admin.chartsFallbackNotice"), "info");
  } catch {
    throw originalError;
  }
}

async function loadChart() {
  clearNotice("pageError");
  state.loading = true;
  state.fallback = false;
  state.metricStats = null;
  state.userStats = null;
  renderAll();

  try {
    const data = await fetchJson(
      `/admin/stats/timeseries?metric=${encodeURIComponent(state.metric)}&days=${state.days}`,
      { headers: authHeaders() },
    );
    state.points = Array.isArray(data?.points) ? data.points : [];
    state.metricStats = data?.metric_stats || null;
    state.userStats = state.metric === "users" ? data?.user_stats || null : null;
    state.granularity = data?.granularity || "daily";
    state.loading = false;
    renderAll();
  } catch (error) {
    state.points = [];
    state.loading = false;
    renderAll();

    if (error?.status === 401 || error?.status === 403) {
      redirectToLogin();
      return;
    }

    if (error?.status === 404) {
      await loadFallbackChartData(error);
      return;
    }

    throw error;
  }
}

function bindControls() {
  document.querySelectorAll("#subscriptionMetricSwitcher .segmented-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      state.metric = METRIC_META[btn.dataset.subscriptionMetric] ? btn.dataset.subscriptionMetric : "subscriptions";
      setQueryParam("metric", state.metric);
      renderControls();
      loadChart().catch(showPageError);
    });
  });

  document.querySelectorAll("#periodSwitcher .segmented-btn").forEach((btn) => {
    btn.addEventListener("click", () => {
      state.days = Number(btn.dataset.days) || 30;
      setQueryParam("days", state.days);
      renderControls();
      loadChart().catch(showPageError);
    });
  });
}

window.addEventListener("DOMContentLoaded", async () => {
  if (!requireAdmin()) return;

  bindLogout();
  highlightActiveNav();

  state.metric = METRIC_META[getQueryParam("metric", "items")] ? getQueryParam("metric", "items") : "items";
  state.days = Number(getQueryParam("days", "30")) || 30;

  bindControls();
  window.addEventListener("trendflix:languagechange", renderAll);

  try {
    await loadChart();
  } catch (error) {
    showPageError(error);
  }
});
