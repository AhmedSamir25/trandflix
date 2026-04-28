const LANG_KEY = "trendflix.lang";

const translations = {
  en: {
    loading: {
      title: "TrendFlix",
      status: "Loading...",
    },
    common: {
      english: "English",
      arabic: "Arabic",
      language: "Language",
      email: "Email",
      name: "Name",
      password: "Password",
      confirmPassword: "Confirm Password",
      togglePasswordVisibility: "Toggle password visibility",
      searchPlaceholder: "Search...",
      watch: "Watch",
      play: "Play",
      read: "Read",
      download: "Download",
      sendMessage: "Send message",
    },
    getStarted: {
      title: "TrendFlix · Get Started",
      subtitle: "Movies, series, games and more",
      emailPlaceholder: "Enter your email",
      submit: "Get Started",
    },
    auth: {
      loginTitle: "TrendFlix · Login",
      signupTitle: "TrendFlix · Sign Up",
      loginSubtitle: "Sign in to continue",
      signupSubtitle: "Create your account",
      login: "Login",
      loggingIn: "Logging in...",
      signup: "Sign Up",
      signingUp: "Signing up...",
      noAccount: "Don't have an account?",
      haveAccount: "Already have an account?",
      signIn: "Sign in",
      signUpLink: "Sign up",
      emailPasswordRequired: "Email and password are required.",
      loginFailed: "Login failed",
      signupFailed: "Signup failed",
      noToken: "No token returned from server",
      allFieldsRequired: "All fields are required.",
      passwordsDoNotMatch: "Passwords do not match.",
      passwordMinLength: "Password must be at least 6 characters.",
    },
    admin: {
      title: "TrendFlix · Admin Dashboard",
      eyebrow: "Admin Dashboard",
      titleHeading: "Manage catalog content",
      subtitle: "Create categories, upload item artwork, and publish new movies, games, and books.",
      backToApp: "Back to app",
      logout: "Logout",
      categoryKicker: "Categories",
      categoryTitle: "Create a category",
      categoryName: "Category name",
      categorySlug: "Category slug",
      createCategory: "Create category",
      creatingCategory: "Creating category...",
      currentCategories: "Current categories",
      itemKicker: "Items",
      itemTitle: "Create a new item",
      itemName: "Title",
      itemType: "Type",
      typeMovie: "Movie",
      typeGame: "Game",
      typeBook: "Book",
      itemDescription: "Description",
      releaseDate: "Release date",
      rating: "Rating",
      coverImage: "Cover image URL",
      uploadImage: "Upload image",
      compressingImage: "Compressing image...",
      author: "Author",
      pagesCount: "Pages count",
      director: "Director",
      duration: "Duration (minutes)",
      developer: "Developer",
      platform: "Platform",
      contentLinkMovie: "Watch link",
      contentLinkMoviePlaceholder: "https://example.com/watch",
      contentLinkGame: "Play link",
      contentLinkGamePlaceholder: "https://example.com/play",
      contentLinkBook: "Read link",
      contentLinkBookPlaceholder: "https://example.com/read",
      assignCategories: "Assign categories",
      createItem: "Create item",
      creatingItem: "Creating item...",
      editItem: "Edit item",
      editItemPageTitle: "TrendFlix · Edit Item",
      editItemHeading: "Edit Item",
      editItemTitle: "Update item details",
      editItemSubtitle: "Update an existing movie, game, or book in the catalog.",
      saveChanges: "Save changes",
      savingChanges: "Saving changes...",
      refresh: "Refresh",
      catalogKicker: "Catalog",
      recentItems: "Recent items",
      noCategories: "No categories yet.",
      noItems: "No items yet.",
      uploadingImage: "Uploading image...",
      imageUploaded: "Image uploaded successfully.",
      adminOnly: "This page is only available to admins.",
      categoryNameRequired: "Category name is required.",
      itemTitleRequired: "Title, type, and release date are required.",
      categoryCreated: "Category created successfully.",
      itemCreated: "Item created successfully.",
      itemUpdated: "Item updated successfully.",
      invalidItemId: "Invalid item id.",
      itemLoadFailed: "Unable to load item details.",
      navDashboard: "Dashboard",
      navCategories: "Categories",
      navCreateItem: "Create Item",
      navCatlog: "Catalog",
      navCatalog: "Catalog",
      categoriesPageTitle: "TrendFlix · Categories",
      createItemPageTitle: "TrendFlix · Create Item",
      catalogPageTitle: "TrendFlix · Catalog",
      dashboardSubtitle: "Track your library, publish content, and control TrendFlix from one focused workspace.",
      hubCategoriesDesc: "Create and manage content categories. Assign slugs and organise your library.",
      hubCategoriesLink: "Go to Categories →",
      hubCreateItemTitle: "Create Item",
      hubCreateItemDesc: "Add a new movie, game, or book to the catalog with cover art and metadata.",
      hubCreateItemLink: "Create Item →",
      hubCatalogDesc: "Browse all published items in the library. Review and audit your content.",
      hubCatalogLink: "View Catalog →",
      categoriesSubtitle: "Create and manage content categories for your library.",
      createItemSubtitle: "Add a new movie, game, or book to the catalog.",
      catalogSubtitle: "Browse all published items in the library.",
      allItems: "All items",
      coverImagePlaceholder: "https://example.com/cover.jpg",
      imagePreviewAlt: "Preview",
      noCategoriesCreate: "No categories yet. Create one first.",
      noItemsCreate: "No items yet. Create one.",
      pages: "pp.",
      statTotalItems: "Total items",
      statTotalItemsHint: "Published in the library",
      statUsers: "Users",
      statUsersHint: "Registered accounts",
      statMovies: "Movies",
      statMoviesHint: "Watchable titles",
      statGames: "Games",
      statGamesHint: "Playable entries",
      statBooks: "Books",
      statBooksHint: "Readable content",
      statCategories: "Categories",
      statCategoriesHint: "Content sections",
      statAvgRating: "Average rating",
      statLatestItem: "Latest item",
      controlKicker: "Control center",
      controlTitle: "Site management",
      healthKicker: "Library health",
      healthTitle: "Content overview",
      recentKicker: "Recent activity",
      viewAll: "View all",
      categorySnapshot: "Items by category",
      deleteItem: "Delete item",
      thisItem: "this item",
      confirmDeleteItem: "Delete {name}? This cannot be undone.",
      editCategory: "Edit",
      deleteCategory: "Delete",
      saveCategory: "Save category",
      savingCategory: "Saving category...",
      categoryUpdated: "Category updated successfully.",
      categoryDeleted: "Category deleted successfully.",
      thisCategory: "this category",
      confirmDeleteCategory: "Delete {name}? Items using it will lose this category.",
    },
    app: {
      title: "TrendFlix · Home",
      sidebarNavigation: "Sidebar navigation",
      home: "Home",
      library: "Library",
      watchLater: "Watch Later",
      favorites: "Favorites",
      settings: "Settings",
      adminDashboard: "Admin Dashboard",
      logout: "Logout",
      openMenu: "Open menu",
      myList: "MY LIST ❤️",
      wishlistItems: "Wishlist items",
      discoverEntertainment: "Discover Entertainment",
      movies: "Movies",
      series: "Series",
      games: "Games",
      books: "Books",
      all: "All",
      action: "Action",
      comedy: "Comedy",
      drama: "Drama",
      mystery: "Mystery",
      boys: "Boys",
      girls: "Girls",
      kids: "Kids",
      scientific: "Scientific",
      novel: "Novel",
      selfHelp: "Self-Help",
      openAiChat: "Open AI chat",
      aiChat: "AI chat",
      trendAi: "TrendFlix",
      closeChat: "Close chat",
      chatScope: "Movies, games, and books only",
      chatWelcome: "Welcome to TrendFlix. I only answer questions about movies, games, and books.",
      askAnything: "Ask about movies, games, or books...",
      chatThinking: "TrendFlix is thinking...",
      chatError: "I couldn't answer right now. Please try again.",
      bannerFallbackTitle: "Featured stories are on the way",
      bannerFallbackDescription: "Add an active banner from the admin dashboard to spotlight movies, games, and books here.",
      toggleFavorite: "Toggle favorite",
      loadingCatalog: "Loading catalog...",
      emptyCatalog: "No items are available yet.",
      noItemsFound: "No items match this filter.",
      catalogLoadFailed: "Unable to load the catalog right now.",
    },
    detail: {
      pageTitle: "TrendFlix · Details",
      loading: "Loading...",
      back: "Back",
      notFound: "Item not found.",
      loadFailed: "Failed to load item.",
      description: "Description",
      director: "Director",
      author: "Author",
      developer: "Developer",
      duration: "Duration",
      mins: "min",
      pages: "Pages",
      platform: "Platform",
      releaseDate: "Release Date",
      reviews: "Reviews",
      noReviews: "No reviews yet. Be the first to review!",
      addFavorite: "Add to Favorites",
      removeFavorite: "Remove from Favorites",
      typeMovie: "Movie",
      typeGame: "Game",
      typeBook: "Book",
    },
    favorites: {
      pageTitle: "TrendFlix · Favorites",
      backToHome: "Back to Home",
      kicker: "Your collection",
      heading: "Favorites",
      subtitle: "Movies, games, and books you saved for later.",
      loading: "Loading favorites...",
      empty: "You do not have any favorite items yet.",
      loadFailed: "Unable to load favorites right now.",
      remove: "Remove",
    },
  },
  ar: {
    loading: {
      title: "ترندفليكس",
      status: "جاري التحميل...",
    },
    common: {
      english: "الإنجليزية",
      arabic: "العربية",
      language: "اللغة",
      email: "البريد الإلكتروني",
      name: "الاسم",
      password: "كلمة المرور",
      confirmPassword: "تأكيد كلمة المرور",
      togglePasswordVisibility: "إظهار أو إخفاء كلمة المرور",
      searchPlaceholder: "ابحث...",
      watch: "شاهد",
      play: "العب",
      read: "اقرأ",
      download: "تحميل",
      sendMessage: "إرسال الرسالة",
    },
    getStarted: {
      title: "ترندفليكس · ابدأ الآن",
      subtitle: "أفلام ومسلسلات وألعاب وأكثر",
      emailPlaceholder: "أدخل بريدك الإلكتروني",
      submit: "ابدأ الآن",
    },
    auth: {
      loginTitle: "ترندفليكس · تسجيل الدخول",
      signupTitle: "ترندفليكس · إنشاء حساب",
      loginSubtitle: "سجّل الدخول للمتابعة",
      signupSubtitle: "أنشئ حسابك",
      login: "تسجيل الدخول",
      loggingIn: "جارٍ تسجيل الدخول...",
      signup: "إنشاء حساب",
      signingUp: "جارٍ إنشاء الحساب...",
      noAccount: "ليس لديك حساب؟",
      haveAccount: "لديك حساب بالفعل؟",
      signIn: "تسجيل الدخول",
      signUpLink: "إنشاء حساب",
      emailPasswordRequired: "البريد الإلكتروني وكلمة المرور مطلوبان.",
      loginFailed: "فشل تسجيل الدخول",
      signupFailed: "فشل إنشاء الحساب",
      noToken: "لم يتم إرجاع رمز الدخول من الخادم",
      allFieldsRequired: "جميع الحقول مطلوبة.",
      passwordsDoNotMatch: "كلمتا المرور غير متطابقتين.",
      passwordMinLength: "يجب أن تتكون كلمة المرور من 6 أحرف على الأقل.",
    },
    admin: {
      title: "ترندفليكس · لوحة الإدارة",
      eyebrow: "لوحة الإدارة",
      titleHeading: "إدارة محتوى المنصة",
      subtitle: "أنشئ التصنيفات، وارفع صور العناصر، وأضف أفلاماً وألعاباً وكتباً جديدة.",
      backToApp: "العودة إلى التطبيق",
      logout: "تسجيل الخروج",
      categoryKicker: "التصنيفات",
      categoryTitle: "إنشاء تصنيف",
      categoryName: "اسم التصنيف",
      categorySlug: "المعرف المختصر",
      createCategory: "إنشاء التصنيف",
      creatingCategory: "جارٍ إنشاء التصنيف...",
      currentCategories: "التصنيفات الحالية",
      itemKicker: "العناصر",
      itemTitle: "إضافة عنصر جديد",
      itemName: "العنوان",
      itemType: "النوع",
      typeMovie: "فيلم",
      typeGame: "لعبة",
      typeBook: "كتاب",
      itemDescription: "الوصف",
      releaseDate: "تاريخ الإصدار",
      rating: "التقييم",
      coverImage: "رابط صورة الغلاف",
      uploadImage: "رفع صورة",
      compressingImage: "جارٍ ضغط الصورة...",
      author: "المؤلف",
      pagesCount: "عدد الصفحات",
      director: "المخرج",
      duration: "المدة (بالدقائق)",
      developer: "المطور",
      platform: "المنصة",
      contentLinkMovie: "رابط المشاهدة",
      contentLinkMoviePlaceholder: "https://example.com/watch",
      contentLinkGame: "رابط اللعب",
      contentLinkGamePlaceholder: "https://example.com/play",
      contentLinkBook: "رابط القراءة",
      contentLinkBookPlaceholder: "https://example.com/read",
      assignCategories: "تعيين التصنيفات",
      createItem: "إنشاء العنصر",
      creatingItem: "جارٍ إنشاء العنصر...",
      editItem: "تعديل العنصر",
      editItemPageTitle: "ترندفليكس · تعديل عنصر",
      editItemHeading: "تعديل عنصر",
      editItemTitle: "تحديث بيانات العنصر",
      editItemSubtitle: "حدّث بيانات فيلم أو لعبة أو كتاب موجود في المنصة.",
      saveChanges: "حفظ التغييرات",
      savingChanges: "جارٍ حفظ التغييرات...",
      refresh: "تحديث",
      catalogKicker: "المحتوى",
      recentItems: "أحدث العناصر",
      noCategories: "لا توجد تصنيفات بعد.",
      noItems: "لا توجد عناصر بعد.",
      uploadingImage: "جارٍ رفع الصورة...",
      imageUploaded: "تم رفع الصورة بنجاح.",
      adminOnly: "هذه الصفحة متاحة للمشرفين فقط.",
      categoryNameRequired: "اسم التصنيف مطلوب.",
      itemTitleRequired: "العنوان والنوع وتاريخ الإصدار مطلوبة.",
      categoryCreated: "تم إنشاء التصنيف بنجاح.",
      itemCreated: "تم إنشاء العنصر بنجاح.",
      itemUpdated: "تم تحديث العنصر بنجاح.",
      invalidItemId: "معرف العنصر غير صالح.",
      itemLoadFailed: "تعذر تحميل بيانات العنصر.",
      navDashboard: "لوحة الإدارة",
      navCategories: "التصنيفات",
      navCreateItem: "إضافة عنصر",
      navCatlog: "المحتوى",
      navCatalog: "المحتوى",
      categoriesPageTitle: "ترندفليكس · التصنيفات",
      createItemPageTitle: "ترندفليكس · إضافة عنصر",
      catalogPageTitle: "ترندفليكس · المحتوى",
      dashboardSubtitle: "تابع المكتبة، وانشر المحتوى، وتحكم في ترندفليكس من مساحة إدارة واحدة.",
      hubCategoriesDesc: "أنشئ وأدر تصنيفات المحتوى. خصص المعرفات ونظّم مكتبتك.",
      hubCategoriesLink: "الانتقال إلى التصنيفات ←",
      hubCreateItemTitle: "إضافة عنصر",
      hubCreateItemDesc: "أضف فيلماً أو لعبة أو كتاباً جديداً مع صورة الغلاف والبيانات الوصفية.",
      hubCreateItemLink: "إضافة عنصر ←",
      hubCatalogDesc: "استعرض جميع العناصر المنشورة في المكتبة. راجع المحتوى وراقبه.",
      hubCatalogLink: "عرض المحتوى ←",
      categoriesSubtitle: "أنشئ تصنيفات المحتوى وأدرها في مكتبتك.",
      createItemSubtitle: "أضف فيلماً أو لعبة أو كتاباً جديداً إلى المنصة.",
      catalogSubtitle: "استعرض جميع العناصر المنشورة في المكتبة.",
      allItems: "جميع العناصر",
      coverImagePlaceholder: "https://example.com/cover.jpg",
      imagePreviewAlt: "معاينة",
      noCategoriesCreate: "لا توجد تصنيفات بعد. أنشئ واحداً أولاً.",
      noItemsCreate: "لا توجد عناصر بعد. أنشئ واحداً.",
      pages: "ص.",
      statTotalItems: "إجمالي العناصر",
      statTotalItemsHint: "منشورة في المكتبة",
      statUsers: "المستخدمون",
      statUsersHint: "حسابات مسجلة",
      statMovies: "الأفلام",
      statMoviesHint: "عناوين قابلة للمشاهدة",
      statGames: "الألعاب",
      statGamesHint: "عناصر قابلة للعب",
      statBooks: "الكتب",
      statBooksHint: "محتوى قابل للقراءة",
      statCategories: "التصنيفات",
      statCategoriesHint: "أقسام المحتوى",
      statAvgRating: "متوسط التقييم",
      statLatestItem: "أحدث عنصر",
      controlKicker: "مركز التحكم",
      controlTitle: "إدارة الموقع",
      healthKicker: "حالة المكتبة",
      healthTitle: "نظرة عامة على المحتوى",
      recentKicker: "النشاط الأخير",
      viewAll: "عرض الكل",
      categorySnapshot: "العناصر حسب التصنيف",
      deleteItem: "حذف العنصر",
      thisItem: "هذا العنصر",
      confirmDeleteItem: "هل تريد حذف {name}؟ لا يمكن التراجع عن ذلك.",
      editCategory: "تعديل",
      deleteCategory: "حذف",
      saveCategory: "حفظ التصنيف",
      savingCategory: "جارٍ حفظ التصنيف...",
      categoryUpdated: "تم تحديث التصنيف بنجاح.",
      categoryDeleted: "تم حذف التصنيف بنجاح.",
      thisCategory: "هذا التصنيف",
      confirmDeleteCategory: "هل تريد حذف {name}؟ ستفقد العناصر المرتبطة به هذا التصنيف.",
    },
    app: {
      title: "ترندفليكس · الرئيسية",
      sidebarNavigation: "التنقل الجانبي",
      home: "الرئيسية",
      library: "المكتبة",
      watchLater: "المشاهدة لاحقاً",
      favorites: "المفضلة",
      settings: "الإعدادات",
      adminDashboard: "لوحة الإدارة",
      logout: "تسجيل الخروج",
      openMenu: "فتح القائمة",
      myList: "قائمتي ❤️",
      wishlistItems: "عناصر القائمة المفضلة",
      discoverEntertainment: "اكتشف عالم الترفيه",
      movies: "الأفلام",
      series: "المسلسلات",
      games: "الألعاب",
      books: "الكتب",
      all: "الكل",
      action: "أكشن",
      comedy: "كوميديا",
      drama: "دراما",
      mystery: "غموض",
      boys: "أولاد",
      girls: "بنات",
      kids: "أطفال",
      scientific: "علمي",
      novel: "رواية",
      selfHelp: "تطوير الذات",
      openAiChat: "فتح دردشة الذكاء الاصطناعي",
      aiChat: "دردشة الذكاء الاصطناعي",
      trendAi: "TrendFlix",
      closeChat: "إغلاق الدردشة",
      chatScope: "للأفلام والألعاب والكتب فقط",
      chatWelcome: "أهلاً بك في ترندفليكس. أجيب فقط عن الأسئلة المتعلقة بالأفلام والألعاب والكتب.",
      askAnything: "اسأل عن فيلم أو لعبة أو كتاب...",
      chatThinking: "TrendFlix يفكر الآن...",
      chatError: "تعذر عليّ الرد الآن. حاول مرة أخرى.",
      bannerFallbackTitle: "القصص المميزة في الطريق",
      bannerFallbackDescription: "أضف بانر مفعلاً من لوحة الإدارة لعرض الأفلام والألعاب والكتب هنا.",
      toggleFavorite: "إضافة أو إزالة من المفضلة",
      loadingCatalog: "جارٍ تحميل المحتوى...",
      emptyCatalog: "لا توجد عناصر متاحة حالياً.",
      noItemsFound: "لا توجد عناصر تطابق هذا الفلتر.",
      catalogLoadFailed: "تعذر تحميل المحتوى حالياً.",
    },
    detail: {
      pageTitle: "ترندفليكس · التفاصيل",
      loading: "جارٍ التحميل...",
      back: "رجوع",
      notFound: "العنصر غير موجود.",
      loadFailed: "فشل تحميل العنصر.",
      description: "الوصف",
      director: "المخرج",
      author: "المؤلف",
      developer: "المطور",
      duration: "المدة",
      mins: "دقيقة",
      pages: "الصفحات",
      platform: "المنصة",
      releaseDate: "تاريخ الإصدار",
      reviews: "التقييمات",
      noReviews: "لا توجد تقييمات بعد. كن أول من يقيّم!",
      addFavorite: "إضافة للمفضلة",
      removeFavorite: "إزالة من المفضلة",
      typeMovie: "فيلم",
      typeGame: "لعبة",
      typeBook: "كتاب",
    },
    favorites: {
      pageTitle: "ترندفليكس · المفضلة",
      backToHome: "العودة للرئيسية",
      kicker: "مجموعتك",
      heading: "المفضلة",
      subtitle: "الأفلام والألعاب والكتب التي حفظتها لوقت لاحق.",
      loading: "جارٍ تحميل المفضلة...",
      empty: "لا توجد عناصر مفضلة لديك بعد.",
      loadFailed: "تعذر تحميل المفضلة حالياً.",
      remove: "إزالة",
    },
  },
};

function getStoredLang() {
  return localStorage.getItem(LANG_KEY) === "ar" ? "ar" : "en";
}

function getValue(source, key) {
  return key.split(".").reduce((value, part) => value?.[part], source);
}

let currentLang = getStoredLang();

function t(key) {
  return getValue(translations[currentLang], key) ?? getValue(translations.en, key) ?? key;
}

function translateOrFallback(key, fallback = "") {
  const translated = t(key);
  return translated === key ? fallback : translated;
}

function translatePage() {
  if (document.documentElement.dataset.i18nTitle) {
    document.title = translateOrFallback(document.documentElement.dataset.i18nTitle, document.title);
  }

  document.querySelectorAll("[data-i18n]").forEach((el) => {
    el.textContent = translateOrFallback(el.dataset.i18n, el.textContent || "");
  });

  document.querySelectorAll("[data-i18n-placeholder]").forEach((el) => {
    el.setAttribute(
      "placeholder",
      translateOrFallback(el.dataset.i18nPlaceholder, el.getAttribute("placeholder") || ""),
    );
  });

  document.querySelectorAll("[data-i18n-aria-label]").forEach((el) => {
    el.setAttribute(
      "aria-label",
      translateOrFallback(el.dataset.i18nAriaLabel, el.getAttribute("aria-label") || ""),
    );
  });

  document.querySelectorAll("[data-i18n-label]").forEach((el) => {
    el.setAttribute(
      "aria-label",
      translateOrFallback(el.dataset.i18nLabel, el.getAttribute("aria-label") || ""),
    );
  });

  document.querySelectorAll("[data-i18n-alt]").forEach((el) => {
    el.setAttribute("alt", translateOrFallback(el.dataset.i18nAlt, el.getAttribute("alt") || ""));
  });

  document.querySelectorAll("[data-set-lang]").forEach((btn) => {
    const active = btn.dataset.setLang === currentLang;
    btn.classList.toggle("active", active);
    btn.setAttribute("aria-pressed", String(active));
  });
}

function closeLangMenus() {
  document.querySelectorAll(".lang-menu.open").forEach((menu) => {
    menu.classList.remove("open");
    menu.querySelector("[data-lang-trigger]")?.setAttribute("aria-expanded", "false");
  });
}

function toggleLangMenu(trigger) {
  const menu = trigger.closest(".lang-menu");
  if (!menu) return;

  const isOpen = menu.classList.contains("open");
  closeLangMenus();
  menu.classList.toggle("open", !isOpen);
  trigger.setAttribute("aria-expanded", String(!isOpen));
}

function applyLanguage(lang) {
  currentLang = lang === "ar" ? "ar" : "en";
  document.documentElement.lang = currentLang;
  document.documentElement.dir = currentLang === "ar" ? "rtl" : "ltr";
  translatePage();
}

function setLang(lang) {
  const nextLang = lang === "ar" ? "ar" : "en";
  localStorage.setItem(LANG_KEY, nextLang);
  applyLanguage(nextLang);
  window.dispatchEvent(new CustomEvent("trendflix:languagechange", { detail: { lang: nextLang } }));
}

document.addEventListener("click", (event) => {
  const trigger = event.target.closest("[data-lang-trigger]");
  if (trigger) {
    toggleLangMenu(trigger);
    return;
  }

  const button = event.target.closest("[data-set-lang]");
  if (button) {
    setLang(button.dataset.setLang);
    closeLangMenus();
    return;
  }

  if (!event.target.closest(".lang-menu")) {
    closeLangMenus();
  }
});

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") closeLangMenus();
});

window.TrendFlixI18n = {
  getLang: () => currentLang,
  setLang,
  t,
  translatePage,
};

applyLanguage(currentLang);
window.addEventListener("DOMContentLoaded", translatePage);
