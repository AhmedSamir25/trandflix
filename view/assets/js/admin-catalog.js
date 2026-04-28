let items = [];
let activeFilter = "all";

const TYPE_ICON = { movie: "🎬", game: "🎮", book: "📖" };

function getTypeLabel(type) {
  if (type === "movie") return t("admin.typeMovie") || "Movie";
  if (type === "game")  return t("admin.typeGame")  || "Game";
  return t("admin.typeBook") || "Book";
}

function formatDuration(minutes) {
  if (!minutes) return null;
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return h > 0 ? `${h}h${m > 0 ? " " + m + "m" : ""}` : `${m}m`;
}

function getTypeMeta(item) {
  if (item.type === "movie") return item.director || formatDuration(item.duration) || null;
  if (item.type === "book")  return item.author || null;
  if (item.type === "game")  return item.developer || item.platform || null;
  return null;
}

function renderFilters() {
  const container = document.getElementById("catalogFilters");
  if (!container) return;

  const counts = { all: items.length, movie: 0, game: 0, book: 0 };
  items.forEach((item) => { if (counts[item.type] !== undefined) counts[item.type]++; });

  const filters = [
    { key: "all",   label: t("app.all")    || "All"    },
    { key: "movie", label: t("app.movies") || "Movies" },
    { key: "game",  label: t("app.games")  || "Games"  },
    { key: "book",  label: t("app.books")  || "Books"  },
  ];

  container.innerHTML = filters
    .map(
      ({ key, label }) => `
        <button class="filter-btn${activeFilter === key ? " active" : ""}${key !== "all" ? " filter-" + key : ""}"
                data-filter="${key}">
          ${key !== "all" ? TYPE_ICON[key] + " " : ""}${escapeHtml(label)}
          <span class="filter-count">${counts[key]}</span>
        </button>`,
    )
    .join("");
}

function renderItemList() {
  const list = document.getElementById("itemList");
  if (!list) return;

  const visible = activeFilter === "all" ? items : items.filter((i) => i.type === activeFilter);

  if (!visible.length) {
    list.innerHTML = `<p class="notice">${escapeHtml(t("admin.noItemsCreate"))}</p>`;
    return;
  }

  list.innerHTML = visible
    .map((item) => {
      const rating   = Number(item.rating) || 0;
      const year     = item.release_date ? new Date(item.release_date).getFullYear() : null;
      const typeMeta = getTypeMeta(item);
      const icon     = TYPE_ICON[item.type] || "";
      const cats     = (item.categories || []).map((c) => escapeHtml(c.name)).join(", ");

      return `
        <article class="catalog-card" data-type="${escapeHtml(item.type)}">
          <div class="catalog-card__img-wrap">
            <img src="${escapeHtml(item.cover_image || getFallbackImage(item.title))}"
                 alt="${escapeHtml(item.title)}" loading="lazy" />
            <span class="catalog-card__type">${icon} ${escapeHtml(getTypeLabel(item.type))}</span>
          </div>
          <div class="catalog-card__title">${escapeHtml(item.title)}</div>
          <div class="catalog-card__info">
            <div class="catalog-card__row">
              ${rating > 0 ? `<span class="info-tag rating">★ ${rating}</span>` : ""}
              ${year ? `<span class="info-tag date">${year}</span>` : ""}
            </div>
            ${typeMeta ? `<span class="info-tag">${escapeHtml(typeMeta)}</span>` : ""}
            ${cats ? `<span class="info-tag cats">${escapeHtml(cats)}</span>` : ""}
          </div>
          <div class="catalog-card__actions">
            <a class="text-btn catalog-card__action" href="/pages/admin/edit-item.html?id=${item.id}">${escapeHtml(t("admin.editItem"))}</a>
            <button class="text-btn danger-btn catalog-card__action" type="button" data-delete-item="${item.id}">
              ${escapeHtml(t("admin.deleteItem"))}
            </button>
          </div>
        </article>`;
    })
    .join("");
}

function render() {
  renderFilters();
  renderItemList();
}

async function loadItems() {
  clearNotice("pageError");
  const data = await fetchJson("/items");
  items = Array.isArray(data?.items) ? data.items : [];
  render();
}

async function deleteItem(itemId) {
  const item = items.find((entry) => Number(entry.id) === Number(itemId));
  const itemName = item?.title || t("admin.thisItem");
  if (!window.confirm(t("admin.confirmDeleteItem").replace("{name}", itemName))) return;

  clearNotice("pageError");
  await fetchJson(`/items/${itemId}`, {
    method: "DELETE",
    headers: authHeaders(),
  });

  items = items.filter((entry) => Number(entry.id) !== Number(itemId));
  render();
}

window.addEventListener("DOMContentLoaded", async () => {
  if (!requireAdmin()) return;

  bindLogout();
  highlightActiveNav();

  document.getElementById("refreshItemsBtn")?.addEventListener("click", () =>
    loadItems().catch(showPageError),
  );

  document.getElementById("catalogFilters")?.addEventListener("click", (event) => {
    const btn = event.target.closest("[data-filter]");
    if (!btn) return;
    activeFilter = btn.dataset.filter;
    render();
  });

  document.getElementById("itemList")?.addEventListener("click", (event) => {
    const btn = event.target.closest("[data-delete-item]");
    if (!btn) return;
    deleteItem(btn.dataset.deleteItem).catch(showPageError);
  });

  window.addEventListener("trendflix:languagechange", render);

  try {
    await loadItems();
  } catch (error) {
    showPageError(error);
  }
});
