let dashboardStats = null;

const DASHBOARD_TYPE_META = {
  movie: { labelKey: "admin.typeMovie", fallback: "Movies", icon: "▶" },
  game: { labelKey: "admin.typeGame", fallback: "Games", icon: "◆" },
  book: { labelKey: "admin.typeBook", fallback: "Books", icon: "◫" },
};

function getDashboardTypeLabel(type) {
  const meta = DASHBOARD_TYPE_META[type];
  return meta ? t(meta.labelKey) || meta.fallback : type;
}

function setText(id, value) {
  const el = document.getElementById(id);
  if (el) el.textContent = value;
}

function getTypeCounts() {
  const counts = { movie: 0, game: 0, book: 0 };
  (dashboardStats?.type_counts || []).forEach((entry) => {
    if (counts[entry.type] !== undefined) counts[entry.type] = Number(entry.count) || 0;
  });
  return counts;
}

function renderTypeBars(typeCounts, totalItems) {
  const container = document.getElementById("typeBars");
  if (!container) return;

  const max = Math.max(Number(totalItems) || 0, 1);
  container.innerHTML = ["movie", "game", "book"]
    .map((type) => {
      const value = typeCounts[type] || 0;
      const percent = Math.round((value / max) * 100);
      const meta = DASHBOARD_TYPE_META[type];

      return `
        <div class="type-bar-row type-${type}">
          <div class="type-bar-label">
            <span>${meta.icon}</span>
            <strong>${escapeHtml(getDashboardTypeLabel(type))}</strong>
            <em>${value}</em>
          </div>
          <div class="type-bar-track">
            <span style="width: ${percent}%"></span>
          </div>
        </div>
      `;
    })
    .join("");
}

function renderRecentItems() {
  const container = document.getElementById("recentItemsList");
  if (!container) return;

  const recent = dashboardStats?.recent_items || [];
  if (!recent.length) {
    container.innerHTML = `<p class="notice">${escapeHtml(t("admin.noItemsCreate"))}</p>`;
    return;
  }

  container.innerHTML = recent
    .map((item) => {
      const rating = Number(item.rating) || 0;
      const year = item.release_date ? new Date(item.release_date).getFullYear() : "";
      const cover = item.cover_image || getFallbackImage(item.title);

      return `
        <a class="recent-item-row" href="/pages/admin/edit-item.html?id=${item.id}">
          <img src="${escapeHtml(cover)}" alt="${escapeHtml(item.title)}" loading="lazy" />
          <span>
            <strong>${escapeHtml(item.title)}</strong>
            <small>${escapeHtml(getDashboardTypeLabel(item.type))}${year ? " · " + year : ""}</small>
          </span>
          <em>${rating > 0 ? "★ " + rating.toFixed(1) : "-"}</em>
        </a>
      `;
    })
    .join("");
}

function renderCategorySnapshot() {
  const container = document.getElementById("categorySnapshot");
  if (!container) return;

  const categories = dashboardStats?.category_counts || [];
  if (!categories.length) {
    container.innerHTML = `<p class="notice">${escapeHtml(t("admin.noCategoriesCreate"))}</p>`;
    return;
  }

  container.innerHTML = categories
    .map(
      (category) => `
        <a class="category-snapshot-pill" href="/pages/admin/categories.html">
          <span>
            <strong>${escapeHtml(category.name)}</strong>
            <code>${escapeHtml(category.slug)}</code>
          </span>
          <em>${Number(category.item_count) || 0}</em>
        </a>
      `,
    )
    .join("");
}

function setDashboardLoading() {
  ["statTotalItems", "statUsers", "statCategoriesCard", "statMovies", "statGames", "statBooks", "statCategories", "statAvgRating", "statLatestItem"].forEach((id) => {
    setText(id, "-");
  });
}

function renderDashboard() {
  if (!dashboardStats) {
    setDashboardLoading();
    return;
  }

  const totalItems = Number(dashboardStats.total_items) || 0;
  const typeCounts = getTypeCounts();
  const latestItem = dashboardStats.latest_item?.title || "-";
  const averageRating = Number(dashboardStats.average_rating) || 0;

  setText("statTotalItems", totalItems);
  setText("statUsers", Number(dashboardStats.total_users) || 0);
  setText("statCategoriesCard", Number(dashboardStats.total_categories) || 0);
  setText("statMovies", typeCounts.movie);
  setText("statGames", typeCounts.game);
  setText("statBooks", typeCounts.book);
  setText("statCategories", Number(dashboardStats.total_categories) || 0);
  setText("statAvgRating", averageRating.toFixed(1));
  setText("statLatestItem", latestItem);

  renderTypeBars(typeCounts, totalItems);
  renderRecentItems();
  renderCategorySnapshot();
}

async function loadDashboard() {
  clearNotice("pageError");
  setDashboardLoading();

  try {
    const data = await fetchJson("/admin/stats", {
      headers: authHeaders(),
    });

    dashboardStats = data?.stats || null;
    renderDashboard();
  } catch (error) {
    dashboardStats = null;
    renderDashboard();

    if (error?.status === 401 || error?.status === 403) {
      redirectToLogin();
      return;
    }

    if (error?.status === 404) {
      throw new Error("Admin stats API is not available. Restart the Go server so /admin/stats is registered.");
    }

    throw error;
  }
}

window.addEventListener("DOMContentLoaded", async () => {
  if (!requireAdmin()) return;

  bindLogout();
  highlightActiveNav();

  document.getElementById("refreshDashboardBtn")?.addEventListener("click", () =>
    loadDashboard().catch(showPageError),
  );

  window.addEventListener("trendflix:languagechange", renderDashboard);

  try {
    await loadDashboard();
  } catch (error) {
    showPageError(error);
  }
});
