let allCommunities = [];
let searchQuery = "";
let searchTimer = 0;

function renderAll() {
  const grid = document.getElementById("allGrid");
  if (!grid) return;

  const filtered = searchQuery
    ? allCommunities.filter((c) =>
        (c.name + " " + (c.description || "")).toLowerCase().includes(searchQuery)
      )
    : allCommunities;

  if (!filtered.length) {
    grid.innerHTML = `<div class="empty-state"><h3>${commEscapeHtml(commT("communities.cEmpty"))}</h3><p>${commEscapeHtml(commT("communities.cEmptyHint"))}</p></div>`;
    return;
  }
  grid.innerHTML = filtered.map(commCommunityCard).join("");
}

function renderPopular(items) {
  const grid = document.getElementById("popularGrid");
  if (!grid) return;
  if (!items.length) {
    grid.innerHTML = "";
    return;
  }
  grid.innerHTML = items.map(commCommunityCard).join("");
}

async function loadPopular() {
  try {
    const res = await commFetch("/api/communities/popular?limit=6");
    renderPopular(res?.data?.communities || []);
  } catch (err) {
    document.getElementById("popularGrid").innerHTML = "";
  }
}

async function loadAll() {
  const grid = document.getElementById("allGrid");
  try {
    const res = await commFetch("/api/communities?per_page=50");
    allCommunities = res?.data?.communities || [];
    renderAll();
  } catch (err) {
    grid.innerHTML = `<p class="comm-status">${commEscapeHtml(commT("communities.cFailed"))}</p>`;
  }
}

window.addEventListener("DOMContentLoaded", async () => {
  commRedirectIfUnauth();

  const input = document.getElementById("searchInput");
  if (input) {
    input.addEventListener("input", (e) => {
      clearTimeout(searchTimer);
      const value = e.target.value.trim().toLowerCase();
      searchTimer = setTimeout(() => {
        searchQuery = value;
        renderAll();
      }, 180);
    });
  }

  window.addEventListener("trendflix:languagechange", () => {
    renderAll();
    loadPopular();
  });

  await Promise.all([loadPopular(), loadAll()]);
});
