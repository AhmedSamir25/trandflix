const TOKEN_KEY = "trendflix.token";

function t(key) {
  return window.TrendFlixI18n?.t(key) ?? key;
}

function setError(msg) {
  const el = document.getElementById("errorMsg");
  el.textContent = msg;
  el.hidden = !msg;
}

async function signup(name, email, password) {
  const res = await fetch("/auth/register", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, email, password }),
  });

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const msg = data?.msg || t("auth.signupFailed");
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

window.addEventListener("DOMContentLoaded", () => {
  initPasswordToggles();

  const token = localStorage.getItem(TOKEN_KEY);
  if (token) {
    window.location.replace("/pages/app.html");
    return;
  }

  const form = document.getElementById("signupForm");
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    setError("");

    const fd = new FormData(form);
    const name = String(fd.get("name") || "").trim();
    const email = String(fd.get("email") || "").trim();
    const password = String(fd.get("password") || "").trim();
    const confirmPassword = String(fd.get("confirmPassword") || "").trim();

    if (!name || !email || !password || !confirmPassword) {
      setError(t("auth.allFieldsRequired"));
      return;
    }

    if (password !== confirmPassword) {
      setError(t("auth.passwordsDoNotMatch"));
      return;
    }

    if (password.length < 6) {
      setError(t("auth.passwordMinLength"));
      return;
    }

    const submitBtn = form.querySelector('button[type="submit"]');
    submitBtn.disabled = true;
    submitBtn.textContent = t("auth.signingUp");

    try {
      const tokenValue = await signup(name, email, password);
      localStorage.setItem(TOKEN_KEY, tokenValue);
      window.location.replace("/pages/app.html");
    } catch (err) {
      setError(err?.message || t("auth.signupFailed"));
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = t("auth.signup");
    }
  });
});
