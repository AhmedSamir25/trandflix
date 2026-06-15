let post = null;
let commentTree = [];
let revealedSpoilers = new Set();
let expandedReplies = new Set();
let collapsedReplies = new Set();
let commentsWide = false;
const REPLIES_PREVIEW = 2;

function postIdFromUrl() {
  return Number(new URLSearchParams(window.location.search).get("id")) || 0;
}

function authorInitial(name) {
  return commEscapeHtml(String(name || "U").charAt(0).toUpperCase());
}

function renderPost() {
  const root = document.getElementById("postRoot");
  const p = post;
  const badges = [];
  if (p.is_pinned) badges.push(`<span class="post-badge post-badge-pin">📌 ${commEscapeHtml(commT("communities.pinned"))}</span>`);
  const typeKey = {
    discussion: "communities.typeDiscussion",
    review: "communities.typeReview",
    poll: "communities.typePoll",
    recommendation_request: "communities.typeRecommendation",
  }[p.post_type];
  badges.push(`<span class="post-badge post-badge-type">${commEscapeHtml(typeKey ? commT(typeKey) : (p.post_type || commT("communities.typeDiscussion")))}</span>`);

  const spoilerId = `spoiler-${p.id}`;
  const revealed = revealedSpoilers.has(spoilerId);
  const bodyHtml = revealed
    ? `<div>${commEscapeHtml(p.body || "")}</div>`
    : `<div class="spoiler-wrap">
         <div class="post-detail-body">${commEscapeHtml(p.body || "")}</div>
         <div class="spoiler-overlay">
           ${commEscapeHtml(commT("communities.pSpoilerWarn"))}
           <small>${commEscapeHtml(commT("communities.pSpoilerDesc"))}</small>
           <button type="button" class="comm-btn comm-btn-primary" data-reveal="${spoilerId}">${commEscapeHtml(commT("communities.pShowSpoiler"))}</button>
         </div>
       </div>`;

  const lockedNote = p.is_locked
    ? `<div class="comment-locked-note">${commEscapeHtml(commT("communities.pLocked"))}</div>`
    : "";

  const communityLink = p.community?.slug
    ? `/pages/community.html?slug=${encodeURIComponent(p.community.slug)}`
    : "/pages/communities.html";

  document.getElementById("backLink").href = communityLink;

  root.innerHTML = `
    <article class="post-detail">
      <h1>${commEscapeHtml(p.title)}</h1>
      <div class="post-detail-meta">
        <span>${badges.join("")}</span>
        <span>${commEscapeHtml(commT("communities.by"))} ${commEscapeHtml(p.user?.name || commT("communities.user"))}</span>
        <span>${commEscapeHtml(commFormatDate(p.created_at))}</span>
        <span>${commEscapeHtml(commT("communities.inLabel"))} <a href="${communityLink}">${commEscapeHtml(p.community?.name || commT("communities.communityFallback"))}</a></span>
      </div>
      ${p.is_spoiler ? bodyHtml : `<div class="post-detail-body">${commEscapeHtml(p.body || "")}</div>`}

      <div class="post-actions">
        <div class="vote-bar" data-post-vote="${p.id}">
          <button type="button" data-vote="up" class="${p.my_vote === "up" ? "voted-up" : ""}">▲</button>
          <span class="score">${Number(p.score || 0)}</span>
          <button type="button" data-vote="down" class="${p.my_vote === "down" ? "voted-down" : ""}">▼</button>
        </div>
      </div>
    </article>

    <section class="comments-section ${commentsWide ? "comments-wide" : ""}">
      <div class="comments-headline">
        <h2>${commEscapeHtml(commT("communities.pCommentsCount").replace("{n}", Number(p.comments_count || 0)))}</h2>
        <button type="button" class="comm-btn comm-btn-ghost comment-width-toggle" data-wide-comments>${commEscapeHtml(commT(commentsWide ? "communities.pNormalComments" : "communities.pWideComments"))}</button>
      </div>
      ${lockedNote}
      ${p.is_locked ? "" : `
        <form class="comment-form" id="commentForm">
          <textarea id="commentBody" placeholder="${commEscapeHtml(commT("communities.pCommentPlaceholder"))}" required></textarea>
          <button type="submit" class="comm-btn comm-btn-primary">${commEscapeHtml(commT("communities.pReplyBtn"))}</button>
        </form>`}
      <div id="commentTree" class="comment-tree"></div>
    </section>
  `;

  renderCommentTree();

  // Bind events
  document.querySelectorAll("[data-reveal]").forEach((btn) => {
    btn.addEventListener("click", () => {
      revealedSpoilers.add(btn.dataset.reveal);
      renderPost();
    });
  });

  document.querySelectorAll("[data-post-vote] button").forEach((btn) => {
    btn.addEventListener("click", () => votePost(btn.dataset.vote));
  });

  const form = document.getElementById("commentForm");
  if (form) {
    form.addEventListener("submit", (e) => {
      e.preventDefault();
      addComment();
    });
  }

  document.querySelector("[data-wide-comments]")?.addEventListener("click", () => {
    commentsWide = !commentsWide;
    renderPost();
  });
}

function renderCommentNode(node) {
  const allReplies = node.replies || [];
  const isCollapsed = collapsedReplies.has(node.id);
  const isExpanded = expandedReplies.has(node.id);
  const visibleReplies = isCollapsed ? [] : (isExpanded ? allReplies : allReplies.slice(0, REPLIES_PREVIEW));
  const hiddenCount = allReplies.length - visibleReplies.length;
  const repliesHtml = visibleReplies.map(renderCommentNode).join("");
  let toggleBtn = "";
  if (isCollapsed) {
    toggleBtn = `<button type="button" class="link comment-toggle" data-show-replies="${node.id}">${commEscapeHtml(commT("communities.pShowReplies").replace("{n}", allReplies.length))}</button>`;
  } else if (hiddenCount > 0) {
    toggleBtn = `<button type="button" class="link comment-toggle" data-more="${node.id}">${commEscapeHtml(commT("communities.pMoreReplies").replace("{n}", hiddenCount))}</button>`;
  }
  if (!isCollapsed && allReplies.length) {
    toggleBtn += `<button type="button" class="link comment-toggle" data-collapse-replies="${node.id}">${commEscapeHtml(commT("communities.pHideReplies"))}</button>`;
  }
  return `
    <div class="comment-node" data-comment-id="${node.id}">
      <div class="comment-head">
        <div class="comment-avatar">${authorInitial(node.user?.name)}</div>
        <span class="comment-author">${commEscapeHtml(node.user?.name || commT("communities.user"))}</span>
        <span class="comment-date">${commEscapeHtml(commFormatDate(node.created_at))}</span>
      </div>
      <div class="comment-text">${commEscapeHtml(node.body)}</div>
      <div class="comment-foot">
        <div class="comment-vote">
          <button type="button" data-cvote="up" data-id="${node.id}">▲</button>
          <span class="score">${Number(node.score || 0)}</span>
          <button type="button" data-cvote="down" data-id="${node.id}">▼</button>
        </div>
        <button type="button" class="link" data-reply="${node.id}">${commEscapeHtml(commT("communities.pReplyBtn"))}</button>
      </div>
      ${(repliesHtml || toggleBtn) ? `<div class="comment-replies">${repliesHtml}${toggleBtn}</div>` : ""}
    </div>
  `;
}

function renderCommentTree() {
  const el = document.getElementById("commentTree");
  if (!el) return;
  if (!commentTree.length) {
    el.innerHTML = `<div class="empty-state"><p>${commEscapeHtml(commT("communities.pNoComments"))}</p></div>`;
    return;
  }
  el.innerHTML = commentTree.map(renderCommentNode).join("");

  el.querySelectorAll("[data-cvote]").forEach((btn) => {
    btn.addEventListener("click", () => voteComment(Number(btn.dataset.id), btn.dataset.cvote));
  });
  el.querySelectorAll("[data-reply]").forEach((btn) => {
    btn.addEventListener("click", () => openReply(Number(btn.dataset.reply)));
  });
  el.querySelectorAll("[data-more]").forEach((btn) => {
    btn.addEventListener("click", () => {
      expandedReplies.add(Number(btn.dataset.more));
      renderCommentTree();
    });
  });
  el.querySelectorAll("[data-less]").forEach((btn) => {
    btn.addEventListener("click", () => {
      expandedReplies.delete(Number(btn.dataset.less));
      renderCommentTree();
    });
  });
  el.querySelectorAll("[data-show-replies]").forEach((btn) => {
    btn.addEventListener("click", () => {
      collapsedReplies.delete(Number(btn.dataset.showReplies));
      renderCommentTree();
    });
  });
  el.querySelectorAll("[data-collapse-replies]").forEach((btn) => {
    btn.addEventListener("click", () => {
      collapsedReplies.add(Number(btn.dataset.collapseReplies));
      renderCommentTree();
    });
  });
}

function openReply(parentId) {
  if (post?.is_locked) return;
  const node = document.querySelector(`.comment-node[data-comment-id="${parentId}"]`);
  if (!node) return;
  if (node.querySelector(".reply-inline")) return;

  const wrap = document.createElement("div");
  wrap.className = "comment-form reply-inline";
  wrap.style.marginTop = "10px";
  wrap.innerHTML = `
    <textarea placeholder="${commEscapeHtml(commT("communities.pReplyPlaceholder"))}" required></textarea>
    <button type="button" class="comm-btn comm-btn-primary" data-send="${parentId}">${commEscapeHtml(commT("communities.pReplyBtn"))}</button>
    <button type="button" class="comm-btn comm-btn-ghost" data-cancel-reply>${commEscapeHtml(commT("communities.cancel"))}</button>
  `;
  node.querySelector(".comment-foot")?.after(wrap);

  wrap.querySelector("[data-send]").addEventListener("click", () => {
    const text = wrap.querySelector("textarea").value.trim();
    if (text) addComment(text, parentId);
  });
  wrap.querySelector("[data-cancel-reply]").addEventListener("click", () => wrap.remove());
  wrap.querySelector("textarea").focus();
}

async function votePost(type) {
  try {
    const res = await commFetch(`/api/posts/${post.id}/vote`, {
      method: "POST",
      body: JSON.stringify({ vote_type: type }),
    });
    post = res?.data?.post || post;
    renderPost();
  } catch (err) {
    alert(err.message || commT("communities.pVoteFailed"));
  }
}

async function voteComment(commentId, type) {
  try {
    const res = await commFetch(`/api/comments/${commentId}/vote`, {
      method: "POST",
      body: JSON.stringify({ vote_type: type }),
    });
    const updated = res?.data?.comment;
    if (updated) {
      const flat = flatten(commentTree);
      const idx = flat.findIndex((c) => c.id === commentId);
      if (idx >= 0) {
        flat[idx].score = updated.score;
        renderCommentTree();
      }
    }
  } catch (err) {
    alert(err.message || commT("communities.pVoteFailed"));
  }
}

function flatten(nodes, acc = []) {
  for (const n of nodes || []) {
    acc.push(n);
    flatten(n.replies, acc);
  }
  return acc;
}

async function addComment(body, parentId) {
  const text = (body || document.getElementById("commentBody")?.value || "").trim();
  if (!text) return;
  try {
    await commFetch(`/api/posts/${post.id}/comments`, {
      method: "POST",
      body: JSON.stringify({ body: text, parent_id: parentId || null }),
    });
    await loadComments();
    if (!body) document.getElementById("commentBody").value = "";
  } catch (err) {
    alert(err.message || commT("communities.pCommentFailed"));
  }
}

async function loadComments() {
  try {
    const res = await commFetch(`/api/posts/${post.id}/comments`);
    commentTree = res?.data?.comments || [];
    renderCommentTree();
    if (post) post.comments_count = flatten(commentTree).length;
  } catch (err) {
    console.error(err);
  }
}

async function loadPost() {
  const id = postIdFromUrl();
  if (!id) {
    document.getElementById("postRoot").innerHTML = `<p class="comm-status">${commEscapeHtml(commT("communities.pInvalid"))}</p>`;
    return;
  }
  try {
    const res = await commFetch(`/api/posts/${id}`);
    post = res?.data?.post;
    if (!post) throw new Error(commT("communities.pNotFound"));
    renderPost();
    await loadComments();
  } catch (err) {
    document.getElementById("postRoot").innerHTML = `<p class="comm-status">${commEscapeHtml(err.message)}</p>`;
  }
}

window.addEventListener("DOMContentLoaded", () => {
  commRedirectIfUnauth();
  loadPost();
  window.addEventListener("trendflix:languagechange", () => {
    if (post) renderPost();
  });
});
