function commCommunityCard(c) {
  const href = `/pages/community.html?slug=${encodeURIComponent(c.slug)}`;
  const avatar = commEscapeHtml(c.avatar_image || "");
  const cover = commEscapeHtml(c.cover_image || "");
  return `
    <a class="comm-card" href="${href}">
      <img class="comm-card-cover" src="${cover}" alt="" loading="lazy" onerror="this.style.display='none'" />
      <div class="comm-card-body">
        ${avatar ? `<img class="comm-card-avatar" src="${avatar}" alt="" />` : ""}
        <div class="comm-card-info">
          <h3 class="comm-card-name">${commEscapeHtml(c.name)}</h3>
          <div class="comm-card-cat">${commEscapeHtml(commCategoryLabel(c.category_type))}</div>
          ${c.description ? `<p class="comm-card-desc">${commEscapeHtml(c.description)}</p>` : ""}
        </div>
      </div>
      <div class="comm-card-stats">
        <span>👥 ${Number(c.members_count || 0).toLocaleString()}</span>
        <span>📝 ${Number(c.posts_count || 0).toLocaleString()}</span>
      </div>
    </a>
  `;
}

function commPostCard(post) {
  const href = `/pages/post.html?id=${encodeURIComponent(post.id)}`;
  const badges = [];
  if (post.is_pinned) badges.push(`<span class="post-badge post-badge-pin">📌 ${commEscapeHtml(commT("communities.pinned"))}</span>`);
  const typeKey = {
    discussion: "communities.typeDiscussion",
    review: "communities.typeReview",
    poll: "communities.typePoll",
    recommendation_request: "communities.typeRecommendation",
  }[post.post_type];
  badges.push(`<span class="post-badge post-badge-type">${commEscapeHtml(typeKey ? commT(typeKey) : (post.post_type || commT("communities.typeDiscussion")))}</span>`);
  if (post.is_spoiler) badges.push(`<span class="post-badge post-badge-spoiler">⚠ ${commEscapeHtml(commT("communities.spoiler"))}</span>`);

  const author = commEscapeHtml(post.user?.name || commT("communities.user"));
  const when = commEscapeHtml(commFormatDate(post.created_at));

  return `
    <a class="post-card" href="${href}">
      <div class="post-score-col">
        <span class="arrow">▲</span>
        <span class="score">${Number(post.score || 0)}</span>
        <span class="arrow">▼</span>
      </div>
      <div class="post-body-col">
        <h3 class="post-card-title">${commEscapeHtml(post.title)}</h3>
        <div class="post-card-meta">
          <span>${commEscapeHtml(commT("communities.by"))} ${author}</span>
          <span>${when}</span>
          <span>💬 ${Number(post.comments_count || 0)}</span>
        </div>
        <div>${badges.join("")}</div>
        ${post.body ? `<p class="post-snippet">${commEscapeHtml(post.body)}</p>` : ""}
      </div>
    </a>
  `;
}
