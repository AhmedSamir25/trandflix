const TOKEN_KEY = "trendflix.token";

function t(key) {
  return window.TrendFlixI18n?.t(key) ?? key;
}

function setError(msg) {
  const el = document.getElementById("errorMsg");
  el.textContent = msg;
  el.hidden = !msg;
}

async function login(email, password) {
  const res = await fetch("/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const msg = data?.msg || t("auth.loginFailed");
    throw new Error(msg);
  }

  if (!data?.token) {
    throw new Error(t("auth.noToken"));
  }
  return data.token;
}

function initPasswordToggles() {
  document.querySelectorAll(".password-toggle").forEach((btn) => {
    btn.addEventListener("click", () => {
      const field = btn.closest(".password-field");
      const input = field.querySelector("input");
      const isPassword = input.type === "password";
      input.type = isPassword ? "text" : "password";
      btn.classList.toggle("showing", !isPassword);
    });
  });
}

function getUrlParam(name) {
  const params = new URLSearchParams(window.location.search);
  return params.get(name);
}

window.addEventListener("DOMContentLoaded", () => {
  initPasswordToggles();

  const token = localStorage.getItem(TOKEN_KEY);
  if (token) {
    window.location.replace("/pages/app.html");
    return;
  }

  const emailFromUrl = getUrlParam("email");
  const emailInput = document.querySelector('input[name="email"]');
  if (emailFromUrl && emailInput) {
    emailInput.value = emailFromUrl;
  }

  const form = document.getElementById("loginForm");
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    setError("");

    const fd = new FormData(form);
    const email = String(fd.get("email") || "").trim();
    const password = String(fd.get("password") || "").trim();

    if (!email || !password) {
      setError(t("auth.emailPasswordRequired"));
      return;
    }

    const submitBtn = form.querySelector('button[type="submit"]');
    submitBtn.disabled = true;
    submitBtn.textContent = t("auth.loggingIn");

    try {
      const tokenValue = await login(email, password);
      localStorage.setItem(TOKEN_KEY, tokenValue);
      window.location.replace("/pages/app.html");
    } catch (err) {
      setError(err?.message || t("auth.loginFailed"));
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = t("auth.login");
    }
  });
});
