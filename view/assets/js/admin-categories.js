let categories = [];
let slugTouched = false;

function renderCategoryList() {
  const list = document.getElementById("categoryList");
  if (!list) return;

  if (!categories.length) {
    list.innerHTML = `<p class="notice">${escapeHtml(t("admin.noCategories"))}</p>`;
    return;
  }

  list.innerHTML = categories
    .map(
      (c) =>
        `<span class="category-pill"><strong>${escapeHtml(c.name)}</strong><code>${escapeHtml(c.slug)}</code></span>`,
    )
    .join("");
}

async function loadCategories() {
  clearNotice("pageError");
  const data = await fetchJson("/categories");
  categories = Array.isArray(data?.categories) ? data.categories : [];
  renderCategoryList();
}

async function handleCategorySubmit(event) {
  event.preventDefault();
  clearNotice("pageError");
  clearNotice("categoryStatus");

  const nameInput = document.getElementById("categoryName");
  const slugInput = document.getElementById("categorySlug");
  const name = String(nameInput?.value || "").trim();
  const slug = slugify(slugInput?.value || name);

  if (!name) {
    setNotice("categoryStatus", t("admin.categoryNameRequired"), "error");
    return;
  }

  if (slugInput) slugInput.value = slug;

  setButtonLoading("createCategoryBtn", "admin.creatingCategory", "admin.createCategory", true);

  try {
    const response = await fetchJson("/categories", {
      method: "POST",
      headers: authHeaders({ "Content-Type": "application/json" }),
      body: JSON.stringify({ name, slug }),
    });

    event.target.reset();
    slugTouched = false;
    setNotice("categoryStatus", response?.msg || t("admin.categoryCreated"), "success");
    await loadCategories();
  } catch (error) {
    setNotice("categoryStatus", error.message, "error");
  } finally {
    setButtonLoading("createCategoryBtn", "admin.creatingCategory", "admin.createCategory", false);
  }
}

window.addEventListener("DOMContentLoaded", async () => {
  if (!requireAdmin()) return;

  bindLogout();
  highlightActiveNav();

  document.getElementById("categoryForm")?.addEventListener("submit", handleCategorySubmit);

  document.getElementById("refreshCategoriesBtn")?.addEventListener("click", () =>
    loadCategories().catch(showPageError),
  );

  document.getElementById("categoryName")?.addEventListener("input", (event) => {
    if (slugTouched) return;
    const slugInput = document.getElementById("categorySlug");
    if (slugInput) slugInput.value = slugify(event.target.value);
  });

  document.getElementById("categorySlug")?.addEventListener("input", () => {
    slugTouched = true;
  });

  window.addEventListener("trendflix:languagechange", renderCategoryList);

  try {
    await loadCategories();
  } catch (error) {
    showPageError(error);
  }
});
