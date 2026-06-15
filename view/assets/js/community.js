let community = null;
let activeSort = "hot";
let isMember = false;

function slugFromUrl() {
  const params = new URLSearchParams(window.location.search);
  return params.get("slug") || "";
}

function renderHeader() {
  const root = document.getElementById("communityRoot");
  const c = community;
  const cover = commCoverUrl(c);
  const avatar = commAvatarUrl(c);

  root.innerHTML = `
    ${c.status === "pending" ? `
      <div class="comm-review-banner">
        <span class="comm-review-icon">⏳</span>
        <div class="comm-review-text">
          <strong>${commEscapeHtml(commT("communities.dPendingBadge"))}</strong>
          <span>${commEscapeHtml(commT("communities.dUnderReview"))}</span>
        </div>
      </div>` : ""}
    <div class="comm-detail-cover">
      <img src="${commEscapeHtml(cover)}" alt="" />
    </div>
    <div class="comm-detail-head">
      <img class="comm-detail-avatar" src="${commEscapeHtml(avatar)}" alt="" />
      <div class="comm-detail-titleblock">
        <h1>${commEscapeHtml(c.name)}</h1>
        <div class="comm-detail-meta">
          <span>${commEscapeHtml(commCategoryLabel(c.category_type))}</span>
          <span>👥 ${commEscapeHtml(commT("communities.membersCount").replace("{n}", Number(c.members_count || 0).toLocaleString()))}</span>
          <span>📝 ${commEscapeHtml(commT("communities.postsCount").replace("{n}", Number(c.posts_count || 0).toLocaleString()))}</span>
        </div>
      </div>
      <div class="comm-detail-actions">
        <button id="joinBtn" class="comm-btn ${isMember ? "comm-btn-danger" : "comm-btn-primary"}">
          ${commEscapeHtml(isMember ? commT("communities.dLeave") : commT("communities.dJoin"))}
        </button>
        <a class="comm-btn comm-btn-ghost" href="#" id="newPostBtn">${commEscapeHtml(commT("communities.dNewPost"))}</a>
      </div>
    </div>

    <div class="comm-detail-body">
      <div class="comm-main">
        <div class="comm-sort-tabs">
          <button data-sort="hot" class="${activeSort === "hot" ? "active" : ""}">${commEscapeHtml(commT("communities.dSortHot"))}</button>
          <button data-sort="new" class="${activeSort === "new" ? "active" : ""}">${commEscapeHtml(commT("communities.dSortNew"))}</button>
          <button data-sort="top" class="${activeSort === "top" ? "active" : ""}">${commEscapeHtml(commT("communities.dSortTop"))}</button>
        </div>
        <div id="postsList" class="post-list">
          <p class="comm-status">${commEscapeHtml(commT("communities.dLoadingPosts"))}</p>
        </div>
      </div>

      <aside class="comm-sidebar">
        ${c.description ? `
          <div class="comm-panel">
            <h3>${commEscapeHtml(commT("communities.dAbout"))}</h3>
            <p>${commEscapeHtml(c.description)}</p>
          </div>` : ""}
        ${c.rules ? `
          <div class="comm-panel">
            <h3>${commEscapeHtml(commT("communities.dRules"))}</h3>
            <p>${commEscapeHtml(c.rules)}</p>
          </div>` : ""}
        <div class="comm-panel">
          <h3>${commEscapeHtml(commT("communities.createdBy"))}</h3>
          <p>${commEscapeHtml(community.user?.name || commT("communities.unknown"))}</p>
        </div>
      </aside>
    </div>
  `;

  document.getElementById("joinBtn").addEventListener("click", toggleMembership);
  document.getElementById("newPostBtn").addEventListener("click", (e) => {
    e.preventDefault();
    if (!isMember) {
      alert(commT("communities.dMustJoin"));
      return;
    }
    window.location.href = `/pages/create-post.html?community=${encodeURIComponent(c.id)}`;
  });
  document.querySelectorAll(".comm-sort-tabs button").forEach((btn) => {
    btn.addEventListener("click", () => {
      activeSort = btn.dataset.sort;
      document.querySelectorAll(".comm-sort-tabs button").forEach((b) => b.classList.remove("active"));
      btn.classList.add("active");
      loadPosts();
    });
  });
}

async function toggleMembership() {
  const btn = document.getElementById("joinBtn");
  btn.disabled = true;
  const path = `/api/communities/${community.id}/${isMember ? "leave" : "join"}`;
  try {
    await commFetch(path, { method: "POST" });
    isMember = !isMember;
    community.members_count += isMember ? 1 : -1;
    renderHeader();
  } catch (err) {
    alert(err.message || commT("communities.dMembershipFailed"));
    btn.disabled = false;
  }
}

async function loadCommunity() {
  const slug = slugFromUrl();
  if (!slug) {
    document.getElementById("communityRoot").innerHTML = `<p class="comm-status">${commEscapeHtml(commT("communities.dNone"))}</p>`;
    return;
  }

  try {
    const res = await commFetch(`/api/communities/${encodeURIComponent(slug)}`);
    community = res?.data?.community;
    if (!community) throw new Error("Community not found");

    // Determine membership for the current user
    if (commToken()) {
      try {
        const mres = await commFetch(`/api/communities/${encodeURIComponent(community.id)}/members?per_page=1`);
        const members = mres?.data?.members || [];
        const me = commCurrentUserId();
        isMember = members.some((m) => Number(m.user_id) === me) && me !== 0;
      } catch {
        isMember = false;
      }
    }

    renderHeader();
    await loadPosts();
  } catch (err) {
    document.getElementById("communityRoot").innerHTML = `<p class="comm-status">${commEscapeHtml(err.message)}</p>`;
  }
}

async function loadPosts() {
  const list = document.getElementById("postsList");
  if (!list) return;
  try {
    const res = await commFetch(
      `/api/communities/${community.id}/posts?sort=${encodeURIComponent(activeSort)}&per_page=30`
    );
    const posts = res?.data?.posts || [];
    if (!posts.length) {
      list.innerHTML = `<div class="empty-state"><h3>${commEscapeHtml(commT("communities.dNoPosts"))}</h3><p>${commEscapeHtml(commT("communities.dNoPostsHint"))}</p></div>`;
      return;
    }
    list.innerHTML = posts.map(commPostCard).join("");
  } catch (err) {
    list.innerHTML = `<p class="comm-status">${commEscapeHtml(err.message)}</p>`;
  }
}

window.addEventListener("DOMContentLoaded", () => {
  commRedirectIfUnauth();
  loadCommunity();
  window.addEventListener("trendflix:languagechange", () => {
    if (community) {
      renderHeader();
      loadPosts();
    }
  });
});
