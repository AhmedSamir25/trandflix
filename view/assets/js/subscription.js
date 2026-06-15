const TOKEN_KEY = "trendflix.token";
const FALLBACK_IMAGE_BASE = "https://placehold.co/500x750/0f172a/f8fafc";

function requireAuth() {
  const token = localStorage.getItem(TOKEN_KEY);
  if (!token) {
    window.location.replace("/pages/auth/auth.html");
    return null;
  }
  return token;
}

async function fetchJson(url, options = {}, token = "") {
  const headers = {
    Accept: "application/json",
    ...(options.headers || {}),
  };
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(url, { ...options, headers });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data?.msg || `Request failed: ${res.status}`);
  return data;
}

function t(key) {
  return window.TrendFlixI18n?.t(key) ?? key;
}

function escapeHtml(s) {
  return String(s ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function formatDate(dateString) {
  if (!dateString) return "";
  const date = new Date(dateString);
  if (isNaN(date.getTime())) return "";
  return date.toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric" });
}

function renderPlanCard(plan) {
  return `
    <header class="sub-header">
      <h1 class="brand">TrendFlix</h1>
      <p class="subtitle" data-i18n="subscription.subtitle">Unlimited movies, TV shows, and books.</p>
    </header>

    <div class="plan-card">
      <div class="plan-badge" data-i18n="subscription.onePlan">One Plan</div>
      <h2 class="plan-name">${escapeHtml(plan.name)}</h2>
      <div class="plan-price">
        <span class="plan-currency">${escapeHtml(plan.currency)}</span>
        <span class="plan-amount">${plan.price.toFixed(2)}</span>
        <span class="plan-period">/ ${plan.billing_period_days} ${t("subscription.days")}</span>
      </div>
      <ul class="plan-features">
        <li>🎬 ${t("subscription.watchMovies")}</li>
        <li>📺 ${t("subscription.watchTvShows")}</li>
        <li>📚 ${t("subscription.readBooks")}</li>
      </ul>
    </div>
  `;
}

function renderSubscriptionStatus(subscription, isActive) {
  const statusClass = isActive ? "status-active" : "status-inactive";
  const statusText = isActive
    ? t("subscription.active")
    : t(`subscription.status_${subscription?.status || "expired"}`);
  const expiresAt = subscription?.ends_at ? formatDate(subscription.ends_at) : "";

  let actionHtml = "";
  if (isActive) {
    actionHtml = `<button class="btn danger subscription-cancel-btn" id="cancelBtn" type="button" data-i18n="subscription.cancel">Cancel Subscription</button>`;
  } else if (subscription?.status === "cancelled") {
    actionHtml = `<p class="info-msg" data-i18n="subscription.activeUntil">Your access remains active until the end of your billing period.</p>`;
  }

  return `
    <div class="subscription-status ${statusClass}">
      <div class="status-badge ${statusClass}">${escapeHtml(statusText)}</div>
      ${expiresAt ? `<p class="status-expiry">${t("subscription.expiresOn")}: ${escapeHtml(expiresAt)}</p>` : ""}
      ${actionHtml}
    </div>
  `;
}

function renderCheckoutForm(plan) {
  return `
    <div class="checkout-section" id="checkoutSection">
      <h2 class="checkout-title" data-i18n="subscription.checkoutTitle">Payment Details</h2>
      <p class="checkout-subtitle" data-i18n="subscription.checkoutSubtitle">This is a mock payment. No real charges will be made.</p>
      <form id="checkoutForm" class="form">
        <label class="field">
          <span data-i18n="subscription.cardName">Cardholder Name</span>
          <input name="card_name" type="text" autocomplete="cc-name" required placeholder="John Doe" />
        </label>
        <label class="field">
          <span data-i18n="subscription.cardNumber">Card Number</span>
          <input name="card_number" type="text" autocomplete="cc-number" required placeholder="4111 1111 1111 1111" maxlength="19" />
        </label>
        <div class="field-row">
          <label class="field">
            <span data-i18n="subscription.cardExpiry">Expiry (MM/YY)</span>
            <input name="card_expiry" type="text" required placeholder="MM/YY" maxlength="5" />
          </label>
          <label class="field">
            <span data-i18n="subscription.cardCvc">CVC</span>
            <input name="card_cvc" type="text" autocomplete="cc-csc" required placeholder="123" maxlength="4" />
          </label>
        </div>
        <button class="btn primary" type="submit" id="checkoutBtn" data-i18n="subscription.subscribe">Start Subscription</button>
        <p class="mock-note" data-i18n="subscription.mockNote">Mock payment — no real charges.</p>
        <p id="checkoutMsg" class="error" hidden></p>
      </form>
    </div>
  `;
}

function renderRenewForm(plan, subscription) {
  const isActive = subscription?.status === "active" && new Date(subscription?.ends_at) > new Date();
  const buttonLabel = isActive ? t("subscription.extend") : t("subscription.renew");

  return `
    <div class="checkout-section" id="checkoutSection">
      <h2 class="checkout-title">${isActive ? t("subscription.extendTitle") : t("subscription.renewTitle")}</h2>
      <p class="checkout-subtitle" data-i18n="subscription.checkoutSubtitle">This is a mock payment. No real charges will be made.</p>
      <form id="renewForm" class="form">
        <label class="field">
          <span data-i18n="subscription.cardName">Cardholder Name</span>
          <input name="card_name" type="text" autocomplete="cc-name" required placeholder="John Doe" />
        </label>
        <label class="field">
          <span data-i18n="subscription.cardNumber">Card Number</span>
          <input name="card_number" type="text" autocomplete="cc-number" required placeholder="4111 1111 1111 1111" maxlength="19" />
        </label>
        <div class="field-row">
          <label class="field">
            <span data-i18n="subscription.cardExpiry">Expiry (MM/YY)</span>
            <input name="card_expiry" type="text" required placeholder="MM/YY" maxlength="5" />
          </label>
          <label class="field">
            <span data-i18n="subscription.cardCvc">CVC</span>
            <input name="card_cvc" type="text" autocomplete="cc-csc" required placeholder="123" maxlength="4" />
          </label>
        </div>
        <button class="btn primary" type="submit" id="renewBtn">${escapeHtml(buttonLabel)}</button>
        <p class="mock-note" data-i18n="subscription.mockNote">Mock payment — no real charges.</p>
        <p id="renewMsg" class="error" hidden></p>
      </form>
    </div>
  `;
}

function buildPage(plan, subscription, isActive) {
  let html = "";
  html += renderPlanCard(plan);

  if (subscription) {
    html += renderSubscriptionStatus(subscription, isActive);
  }

  if (!isActive) {
    if (subscription && subscription.status !== "expired") {
      html += renderRenewForm(plan, subscription);
    } else {
      html += renderCheckoutForm(plan);
    }
  } else {
    html += renderRenewForm(plan, subscription);
  }

  html += `
    <p class="back-home">
      <a href="/pages/app.html">← ${escapeHtml(t("subscription.backToHome"))}</a>
    </p>
  `;

  return html;
}

function attachHandlers(plan, isActive, token) {
  const checkoutForm = document.getElementById("checkoutForm");
  if (checkoutForm) {
    checkoutForm.addEventListener("submit", async (e) => {
      e.preventDefault();
      const formData = new FormData(checkoutForm);
      const body = {
        card_name: String(formData.get("card_name") || "").trim(),
        card_number: String(formData.get("card_number") || "").trim(),
        card_expiry: String(formData.get("card_expiry") || "").trim(),
        card_cvc: String(formData.get("card_cvc") || "").trim(),
      };

      const btn = document.getElementById("checkoutBtn");
      const msgEl = document.getElementById("checkoutMsg");
      btn.disabled = true;
      btn.textContent = t("subscription.processing");
      msgEl.hidden = true;

      try {
        const res = await fetch("/subscription/checkout", {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
            Accept: "application/json",
          },
          body: JSON.stringify(body),
        });
        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
          msgEl.textContent = data?.msg || t("subscription.checkoutFailed");
          msgEl.hidden = false;
          return;
        }
        window.location.reload();
      } catch (err) {
        msgEl.textContent = err?.message || t("subscription.checkoutFailed");
        msgEl.hidden = false;
      } finally {
        btn.disabled = false;
        btn.textContent = t("subscription.subscribe");
      }
    });
  }

  const renewForm = document.getElementById("renewForm");
  if (renewForm) {
    renewForm.addEventListener("submit", async (e) => {
      e.preventDefault();
      const formData = new FormData(renewForm);
      const body = {
        card_name: String(formData.get("card_name") || "").trim(),
        card_number: String(formData.get("card_number") || "").trim(),
        card_expiry: String(formData.get("card_expiry") || "").trim(),
        card_cvc: String(formData.get("card_cvc") || "").trim(),
      };

      const btn = document.getElementById("renewBtn");
      const msgEl = document.getElementById("renewMsg");
      btn.disabled = true;
      btn.textContent = t("subscription.processing");
      msgEl.hidden = true;

      try {
        const res = await fetch("/subscription/renew", {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token}`,
            "Content-Type": "application/json",
            Accept: "application/json",
          },
          body: JSON.stringify(body),
        });
        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
          msgEl.textContent = data?.msg || t("subscription.renewFailed");
          msgEl.hidden = false;
          return;
        }
        window.location.reload();
      } catch (err) {
        msgEl.textContent = err?.message || t("subscription.renewFailed");
        msgEl.hidden = false;
      } finally {
        btn.disabled = false;
        btn.textContent = isActive ? t("subscription.extend") : t("subscription.renew");
      }
    });
  }

  const cancelBtn = document.getElementById("cancelBtn");
  if (cancelBtn) {
    cancelBtn.addEventListener("click", async () => {
      if (!window.confirm(t("subscription.confirmCancel"))) return;
      cancelBtn.disabled = true;
      try {
        const res = await fetch("/subscription/cancel", {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token}`,
            Accept: "application/json",
          },
        });
        const data = await res.json().catch(() => ({}));
        if (!res.ok) {
          alert(data?.msg || t("subscription.cancelFailed"));
          return;
        }
        window.location.reload();
      } catch {
        alert(t("subscription.cancelFailed"));
      } finally {
        cancelBtn.disabled = false;
      }
    });
  }
}

window.addEventListener("DOMContentLoaded", async () => {
  const token = requireAuth();
  if (!token) return;

  try {
    const [planRes, subRes] = await Promise.all([
      fetchJson("/subscription/plan", {}, token),
      fetchJson("/subscription/me", {}, token).catch(() => ({ subscription: null, is_active: false })),
    ]);

    const plan = planRes?.plan;
    const subscription = subRes?.subscription || null;
    const isActive = subRes?.is_active || false;

    if (!plan) {
      document.getElementById("subscriptionRoot").innerHTML = `
        <div class="error-state">
          <h1 class="brand">TrendFlix</h1>
          <p data-i18n="subscription.noPlan">No subscription plan is available right now.</p>
          <a href="/pages/app.html" class="back-home-link">← ${escapeHtml(t("subscription.backToHome"))}</a>
        </div>`;
      window.TrendFlixI18n?.translatePage();
      return;
    }

    document.getElementById("subscriptionRoot").innerHTML = buildPage(plan, subscription, isActive);
    window.TrendFlixI18n?.translatePage();
    attachHandlers(plan, isActive, token);
  } catch (err) {
    document.getElementById("subscriptionRoot").innerHTML = `
      <div class="error-state">
        <h1 class="brand">TrendFlix</h1>
        <p>${escapeHtml(err?.message || t("subscription.loadFailed"))}</p>
        <a href="/pages/app.html" class="back-home-link">← ${escapeHtml(t("subscription.backToHome"))}</a>
      </div>`;
    window.TrendFlixI18n?.translatePage();
  }
});
