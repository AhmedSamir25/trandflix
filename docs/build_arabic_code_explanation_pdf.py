from pathlib import Path
import re
import textwrap

from PIL import Image, ImageDraw, ImageFont


ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "docs" / "trendflix_code_explanation_ar_20_pages.pdf"

PAGE_W, PAGE_H = 1240, 1754
MARGIN = 70
CODE_TOP = 185
CODE_H = 880
EXPLAIN_TOP = CODE_TOP + CODE_H + 70

AR_FONT = "/usr/share/fonts/vazirmatn-vf-fonts/Vazirmatn[wght].ttf"
MONO_FONT = "/usr/share/fonts/google-noto/NotoSansMono-Regular.ttf"

title_font = ImageFont.truetype(AR_FONT, 46)
subtitle_font = ImageFont.truetype(AR_FONT, 25)
body_font = ImageFont.truetype(AR_FONT, 31)
small_font = ImageFont.truetype(AR_FONT, 22)
code_font = ImageFont.truetype(MONO_FONT, 18)
code_font_small = ImageFont.truetype(MONO_FONT, 16)
line_font = ImageFont.truetype(MONO_FONT, 16)
symbol_font = ImageFont.truetype("/usr/share/fonts/gdouros-symbola/Symbola.ttf", 18)
symbol_font_small = ImageFont.truetype("/usr/share/fonts/gdouros-symbola/Symbola.ttf", 16)

GO_KEYWORDS = {
    "package", "import", "func", "return", "if", "else", "var", "const", "type",
    "struct", "map", "nil", "true", "false", "for", "range", "switch", "case",
    "default", "defer", "go", "select", "interface", "error",
}
JS_KEYWORDS = {
    "const", "let", "var", "function", "return", "if", "else", "for", "of",
    "in", "async", "await", "try", "catch", "finally", "new", "throw", "true",
    "false", "null", "undefined", "class", "import", "export",
}


PAGES = [
    {
        "title": "تشغيل التطبيق وربط المكونات",
        "file": "main.go",
        "ranges": [(12, 39)],
        "explanation": (
            "هذه الدالة هي نقطة بداية مشروع TrendFlix. في البداية يتم الاتصال بقاعدة البيانات، "
            "ثم تنفيذ الـ migration وإنشاء بيانات أولية مثل المدير والتصنيفات والعناصر. بعد ذلك "
            "يتم إنشاء تطبيق Fiber، فتح مسار ملفات الرفع، وتسجيل كل مجموعات الـ routes الخاصة "
            "بالتسجيل، العناصر، المفضلة، التقييمات، الشات، والواجهات. في النهاية يقرأ التطبيق "
            "رقم المنفذ من APP_PORT، وإذا لم يكن موجودًا يعمل افتراضيًا على المنفذ 4000."
        ),
    },
    {
        "title": "الاتصال بقاعدة البيانات",
        "file": "database/db_conn.go",
        "ranges": [(19, 39), (42, 52)],
        "explanation": (
            "هذا الجزء مسؤول عن تجهيز اتصال MySQL باستخدام GORM. الكود يحاول قراءة ملف .env، "
            "ثم يبني DSN من متغيرات البيئة إذا لم يكن DB_DSN جاهزًا. بعد ذلك يفتح الاتصال، "
            "يتأكد من أن handle الخاص بقاعدة البيانات صالح، ويجرب Ping للتأكد أن السيرفر متاح. "
            "عند نجاح كل ذلك يتم تخزين الاتصال في DbConn ليستخدمه باقي المشروع."
        ),
    },
    {
        "title": "نموذج العنصر داخل النظام",
        "file": "models/item_model.go",
        "ranges": [(5, 25)],
        "explanation": (
            "Model Item يمثل أي محتوى داخل TrendFlix سواء كان فيلمًا أو لعبة أو كتابًا. يحتوي على "
            "العنوان، الوصف، النوع، صورة الغلاف، رابط المحتوى، تاريخ الإصدار، وحقول تختلف حسب النوع "
            "مثل المؤلف للكتب أو المخرج للأفلام أو المطور للألعاب. وجود Categories بعلاقة many-to-many "
            "يسمح للعنصر الواحد أن ينتمي لأكثر من تصنيف."
        ),
    },
    {
        "title": "إنشاء مستخدم جديد",
        "file": "controller/auth/auth_controller.go",
        "ranges": [(66, 88), (94, 103)],
        "explanation": (
            "هذا الجزء من CreateUser يعالج عملية التسجيل. يتم تنظيف الاسم والبريد وكلمة المرور، ثم "
            "التحقق أن البيانات الأساسية موجودة. بعد ذلك يبحث في قاعدة البيانات لمنع تكرار البريد. "
            "إذا كان البريد جديدًا، يتم تشفير كلمة المرور باستخدام bcrypt قبل الحفظ، ثم إنشاء سجل "
            "المستخدم بدور user. هذه الخطوات تحمي كلمة المرور وتمنع إنشاء حسابات مكررة بنفس البريد."
        ),
    },
    {
        "title": "تسجيل الدخول وإنشاء JWT",
        "file": "controller/auth/auth_controller.go",
        "ranges": [(145, 160), (189, 202)],
        "explanation": (
            "LoginUser يستقبل البريد وكلمة المرور، ينظف المدخلات، ثم يبحث عن المستخدم في قاعدة البيانات. "
            "إذا لم يوجد المستخدم أو كانت كلمة المرور غير صحيحة يرجع Unauthorized. عند نجاح المطابقة "
            "يتم إنشاء JWT يحتوي على رقم المستخدم والدور ووقت الانتهاء. هذا التوكن هو ما يستخدمه "
            "الفرونت إند لاحقًا للوصول إلى الصفحات والعمليات المحمية."
        ),
    },
    {
        "title": "التحقق من التوكن وحماية الطلبات",
        "file": "middleware/auth_middleware.go",
        "ranges": [(30, 43), (45, 52), (74, 83)],
        "explanation": (
            "Middleware Authenticate يحمي الـ APIs التي تحتاج مستخدمًا مسجلًا. يقرأ Authorization header "
            "ويتأكد أن الصيغة Bearer token صحيحة، ثم يستخدم JWT_SECRET للتحقق من صحة التوقيع. بعد ذلك "
            "يستخرج user id من claim sub ويبحث عن المستخدم في قاعدة البيانات. عند النجاح يخزن المستخدم "
            "في Locals باسم currentUser حتى تستخدمه الـ controllers التالية."
        ),
    },
    {
        "title": "صلاحيات المدير ومسارات العناصر",
        "file": "routers/item_router.go",
        "ranges": [(10, 17)],
        "explanation": (
            "هذا الـ router يقسم عمليات العناصر إلى قسمين. عرض كل العناصر أو عنصر واحد متاح بدون حماية، "
            "لأن المستخدم يحتاج تصفح الكتالوج. أما إنشاء عنصر أو تعديله أو حذفه فيمر عبر Authenticate "
            "ثم RequireAdmin، لذلك لا يستطيع تنفيذ هذه العمليات إلا مستخدم مسجل ودوره admin."
        ),
    },
    {
        "title": "جلب قائمة العناصر من قاعدة البيانات",
        "file": "controller/item_controller/item_controller.go",
        "ranges": [(34, 57)],
        "explanation": (
            "GetItems يرجع قائمة العناصر كاملة مرتبة من الأحدث إلى الأقدم، ويستخدم Preload لتحميل "
            "التصنيفات المرتبطة بكل عنصر. قبل الاستعلام يتحقق أن اتصال قاعدة البيانات موجود، وإذا حدث "
            "خطأ في الاستعلام يرجع استجابة JSON بحالة فشل ورسالة Database error. عند النجاح يضع "
            "العناصر داخل context باسم items ويرجعها للواجهة."
        ),
    },
    {
        "title": "إنشاء عنصر باستخدام Transaction",
        "file": "controller/item_controller/item_controller.go",
        "ranges": [(119, 132), (135, 149)],
        "explanation": (
            "CreateItem مسؤول عن إضافة فيلم أو لعبة أو كتاب جديد. بعد التأكد من وجود اتصال وقبول المدير، "
            "يقرأ جسم الطلب ثم يبني العنصر ويتحقق من التصنيفات. استخدام transaction مهم هنا لأن إنشاء "
            "العنصر وربطه بالتصنيفات يجب أن ينجحا معًا. إذا فشلت أي خطوة يتم Rollback، وإذا نجحت كل "
            "الخطوات يتم Commit وإرجاع العنصر الجديد."
        ),
    },
    {
        "title": "تنظيف بيانات العنصر حسب النوع",
        "file": "controller/item_controller/item_controller.go",
        "ranges": [(381, 401), (419, 429)],
        "explanation": (
            "buildItemFromRequest تحول بيانات الطلب إلى Model صالح للحفظ. الكود ينظف النصوص، يتحقق من "
            "الحقول المطلوبة، يقبل فقط الأنواع book أو movie أو game، ويتأكد من صيغة التاريخ. بعد إنشاء "
            "العنصر يستخدم switch لإزالة الحقول غير المناسبة لكل نوع؛ مثل حذف director من الكتاب أو "
            "author من الفيلم. هذا يحافظ على بيانات منظمة داخل قاعدة البيانات."
        ),
    },
    {
        "title": "إنشاء تصنيف جديد",
        "file": "controller/categories_controller/categories_controller.go",
        "ranges": [(54, 75), (78, 88)],
        "explanation": (
            "CreateCategory يستقبل name و slug من الطلب، ثم ينظف القيم ويتأكد أن الحقلين موجودان. "
            "بعد ذلك يبحث عن slug مكرر لأن الرابط المختصر للتصنيف يجب أن يكون فريدًا. إذا وجد تكرارًا "
            "يرجع Conflict، أما إذا لم يوجد فيكمل حفظ التصنيف الجديد في قاعدة البيانات وإرجاع بياناته."
        ),
    },
    {
        "title": "قائمة المفضلة للمستخدم",
        "file": "controller/favorites_controller/favorites_controller.go",
        "ranges": [(16, 49)],
        "explanation": (
            "GetFavorites تعرض العناصر التي أضافها المستخدم الحالي إلى المفضلة. الدالة تحصل على المستخدم "
            "من currentUser الذي وضعه الـ middleware، ثم تنفذ JOIN بين جدول items وجدول favorites بناءً "
            "على user_id. كذلك يتم تحميل التصنيفات وترتيب النتائج حسب تاريخ الإضافة للمفضلة من الأحدث "
            "إلى الأقدم."
        ),
    },
    {
        "title": "إضافة عنصر إلى المفضلة",
        "file": "controller/favorites_controller/favorites_controller.go",
        "ranges": [(84, 106)],
        "explanation": (
            "AddFavorite تضيف عنصرًا للمفضلة بطريقة آمنة. أولًا تتأكد من المستخدم الحالي ومن رقم العنصر، "
            "ثم تتحقق أن العنصر موجود فعلًا. بعد ذلك تبحث هل نفس المستخدم أضاف نفس العنصر سابقًا؛ إذا "
            "كان موجودًا ترجع Conflict بدل تكرار السجل. إذا لم يكن موجودًا تنشئ سجل Favorite يحتوي "
            "على user id و item id."
        ),
    },
    {
        "title": "إنشاء تقييم للمحتوى",
        "file": "controller/reviews_controller/reviews_controller.go",
        "ranges": [(77, 104)],
        "explanation": (
            "CreateReview يسمح للمستخدم بتقييم عنصر معين. الكود يقرأ الطلب، يتحقق أن التقييم بين 1 و5 "
            "وأن item_id موجود، ثم يتأكد أن العنصر موجود في قاعدة البيانات. بعد ذلك يمنع المستخدم من "
            "تقييم نفس العنصر أكثر من مرة. عند نجاح الشروط يربط التقييم بالمستخدم الحالي ويحفظه."
        ),
    },
    {
        "title": "خدمة الشات داخل TrendFlix",
        "file": "controller/chat_controller/chat_controller.go",
        "ranges": [(62, 89)],
        "explanation": (
            "Reply هي نقطة API الخاصة بالمساعد الذكي داخل التطبيق. تتحقق أولًا من وجود OPENROUTER_API_KEY، "
            "ثم تقرأ رسالة المستخدم وتمنع الرسائل الفارغة أو الطويلة جدًا. بعد ذلك تبني رسائل النظام "
            "التي تحدد أن المساعد يجيب فقط عن الأفلام والألعاب والكتب، وتضيف سجل المحادثة والرسالة "
            "الجديدة قبل تجهيز طلب OpenRouter."
        ),
    },
    {
        "title": "تنظيم سجل الشات واللغة",
        "file": "controller/chat_controller/chat_controller.go",
        "ranges": [(174, 191), (216, 225)],
        "explanation": (
            "normalizeHistory يقلل سجل المحادثة إلى آخر عدد محدد من الرسائل، ويتجاهل أي role غير user "
            "أو assistant، ويقص النص الطويل. هذا يمنع إرسال بيانات كثيرة إلى خدمة الذكاء الاصطناعي. "
            "أما buildLanguageInstruction و isArabicText فيحددان لغة الرد حسب آخر رسالة؛ فإذا كانت "
            "عربية يطلب من المساعد الرد بالعربية، وإلا يرد بالإنجليزية."
        ),
    },
    {
        "title": "رفع الصور والتحقق من نوع الملف",
        "file": "controller/upload_controller/upload_controller.go",
        "ranges": [(33, 62)],
        "explanation": (
            "uploadImage يستخدم لرفع صورة المستخدم أو صورة العنصر. يستقبل الملف من form field باسم file، "
            "ثم يتحقق من الامتداد ونوع المحتوى حتى يقبل الصور فقط. بعد ذلك ينشئ مجلد التخزين إذا لم "
            "يكن موجودًا، يبني اسم ملف جديد باستخدام الوقت الحالي لتجنب التصادم، يحفظ الملف، ثم يرجع "
            "المسار العام الذي يستخدمه الفرونت إند لعرض الصورة."
        ),
    },
    {
        "title": "عرض صفحات الواجهة والملفات الثابتة",
        "file": "routers/view_router.go",
        "ranges": [(10, 31)],
        "explanation": (
            "RegisterViewRoutes يربط صفحات HTML والملفات الثابتة بتطبيق Fiber. الصفحة الرئيسية ترسل "
            "index.html، ويوجد redirect لمسار قديم خاص بصفحة تسجيل الدخول للحفاظ على التوافق. مسار "
            "detail/:id يعرض صفحة التفاصيل، بينما Static يفتح مجلد assets لملفات CSS و JavaScript "
            "ومجلد pages لباقي صفحات الواجهة."
        ),
    },
    {
        "title": "فلترة الكتالوج وبناء كروت العرض",
        "file": "view/assets/js/app.js",
        "ranges": [(103, 113), (128, 147)],
        "explanation": (
            "في الواجهة الأمامية، getFilteredItems تطبق فلترة النوع والتصنيف والبحث النصي على العناصر. "
            "دالة card تبني HTML لكارت واحد يحتوي على الصورة والعنوان والتقييم والتاريخ والتصنيفات "
            "وزر المفضلة. هذه الدوال هي الأساس الذي يجعل الكتالوج قابلًا للبحث والتصفية، ويحول بيانات "
            "الـ API إلى عناصر مرئية يستطيع المستخدم فتح تفاصيلها أو إضافتها للمفضلة."
        ),
    },
    {
        "title": "بناء صفحة تفاصيل العنصر",
        "file": "view/assets/js/detail.js",
        "ranges": [(68, 76), (117, 136)],
        "explanation": (
            "detail.js يحول بيانات العنصر إلى صفحة تفاصيل كاملة. getTypeMeta تحدد النص والأيقونة وزر "
            "الإجراء حسب النوع: مشاهدة، قراءة، أو لعب. buildMeta و buildCategories و buildReviews تنسق "
            "بيانات إضافية مثل المخرج أو المؤلف والتصنيفات والتقييمات. buildPage يجمع كل هذه الأجزاء "
            "في HTML واحد يعرض الخلفية، الصورة، العنوان، التقييم، الأزرار، والوصف."
        ),
    },
]


def read_ranges(file_name, ranges):
    path = ROOT / file_name
    lines = path.read_text(encoding="utf-8").splitlines()
    out = []
    for index, (start, end) in enumerate(ranges):
        if index:
            out.append((None, "    ..."))
        for line_no in range(start, min(end, len(lines)) + 1):
            out.append((line_no, lines[line_no - 1].expandtabs(4)))
    return out


def language_for(file_name):
    if file_name.endswith(".go"):
        return "go"
    if file_name.endswith(".js"):
        return "js"
    return "text"


def split_code_line(line, max_chars=105):
    if len(line) <= max_chars:
        return [line]
    indent = len(line) - len(line.lstrip(" "))
    wrapped = textwrap.wrap(
        line,
        width=max_chars,
        replace_whitespace=False,
        drop_whitespace=False,
        subsequent_indent=" " * min(indent + 4, 16),
    )
    return wrapped or [line[:max_chars]]


def tokens_for(segment, lang):
    keywords = GO_KEYWORDS if lang == "go" else JS_KEYWORDS
    parts = []
    pos = 0

    comment_match = re.search(r"//.*", segment)
    comment_start = comment_match.start() if comment_match else len(segment)
    code_part = segment[:comment_start]
    comment_part = segment[comment_start:]

    pattern = re.compile(r"\"(?:\\.|[^\"])*\"|'(?:\\.|[^'])*'|`[^`]*`|\b[A-Za-z_][A-Za-z0-9_]*\b|\d+(?:\.\d+)?")
    for match in pattern.finditer(code_part):
        if match.start() > pos:
            parts.append((code_part[pos:match.start()], "#d6deeb"))
        token = match.group(0)
        if token in keywords:
            color = "#82aaff"
        elif token.startswith(("\"", "'", "`")):
            color = "#ecc48d"
        elif token[0].isdigit():
            color = "#f78c6c"
        else:
            color = "#d6deeb"
        parts.append((token, color))
        pos = match.end()
    if pos < len(code_part):
        parts.append((code_part[pos:], "#d6deeb"))
    if comment_part:
        parts.append((comment_part, "#637777"))
    return parts


def rounded_rectangle(draw, xy, radius, fill, outline=None, width=1):
    draw.rounded_rectangle(xy, radius=radius, fill=fill, outline=outline, width=width)


def draw_code_panel(draw, page, item):
    left = MARGIN
    top = CODE_TOP
    right = PAGE_W - MARGIN
    bottom = top + CODE_H
    rounded_rectangle(draw, (left, top, right, bottom), 24, "#101720", "#223044", 2)

    header_h = 68
    draw.rounded_rectangle((left, top, right, top + header_h), 24, fill="#172230")
    draw.rectangle((left, top + header_h - 24, right, top + header_h), fill="#172230")

    for i, color in enumerate(["#ff5f57", "#febc2e", "#28c840"]):
        draw.ellipse((left + 28 + i * 34, top + 24, left + 48 + i * 34, top + 44), fill=color)

    draw.text((left + 145, top + 19), item["file"], font=small_font, fill="#cbd5e1")

    code_lines = []
    for line_no, line in read_ranges(item["file"], item["ranges"]):
        split_lines = split_code_line(line)
        for idx, part in enumerate(split_lines):
            code_lines.append((line_no if idx == 0 else None, part))

    usable_h = CODE_H - header_h - 48
    line_h = 25
    max_lines = usable_h // line_h
    font = code_font if len(code_lines) <= max_lines else code_font_small
    line_h = 25 if font is code_font else 22
    max_lines = usable_h // line_h
    if len(code_lines) > max_lines:
        code_lines = code_lines[: max_lines - 1] + [(None, "    ...")]

    gutter_x = left + 35
    code_x = left + 108
    y = top + header_h + 26
    lang = language_for(item["file"])

    for line_no, line in code_lines:
        line_no_text = "" if line_no is None else str(line_no)
        draw.text((gutter_x, y), line_no_text.rjust(4), font=line_font, fill="#52627a")
        x = code_x
        for text, color in tokens_for(line, lang):
            x = draw_code_segment(draw, x, y, text, font, color)
        y += line_h


def draw_code_segment(draw, x, y, text, font, color):
    fallback = symbol_font if font is code_font else symbol_font_small
    for char in text:
        char_font = fallback if ord(char) > 127 else font
        draw.text((x, y), char, font=char_font, fill=color)
        x += draw.textlength(char, font=char_font)
    return x


def wrap_arabic(draw, text, font, max_width):
    words = text.split()
    lines = []
    line = ""
    for word in words:
        candidate = word if not line else line + " " + word
        width = draw.textbbox((0, 0), candidate, font=font, direction="rtl", language="ar")[2]
        if width <= max_width:
            line = candidate
        else:
            if line:
                lines.append(line)
            line = word
    if line:
        lines.append(line)
    return lines


def draw_arabic_block(draw, x_right, y, text, font, fill, max_width, line_gap=12):
    for line in wrap_arabic(draw, text, font, max_width):
        draw.text((x_right, y), line, font=font, fill=fill, direction="rtl", language="ar", anchor="ra")
        y += font.size + line_gap
    return y


def draw_page(index, item, total):
    img = Image.new("RGB", (PAGE_W, PAGE_H), "#f5f7fb")
    draw = ImageDraw.Draw(img)

    draw.rectangle((0, 0, PAGE_W, 128), fill="#0f172a")
    draw.text((PAGE_W - MARGIN, 44), item["title"], font=title_font, fill="#ffffff", direction="rtl", language="ar", anchor="ra")
    draw.text((MARGIN, 54), f"TrendFlix Code Explanation  {index:02d}/{total}", font=subtitle_font, fill="#a7f3d0")

    draw_code_panel(draw, img, item)

    explain_left = MARGIN
    explain_right = PAGE_W - MARGIN
    explain_bottom = PAGE_H - 105
    rounded_rectangle(draw, (explain_left, EXPLAIN_TOP, explain_right, explain_bottom), 22, "#ffffff", "#dde6f3", 2)
    draw.text((explain_right - 28, EXPLAIN_TOP + 36), "شرح الجزء:", font=body_font, fill="#0f172a", direction="rtl", language="ar", anchor="ra")
    draw_arabic_block(
        draw,
        explain_right - 28,
        EXPLAIN_TOP + 95,
        item["explanation"],
        body_font,
        "#243044",
        explain_right - explain_left - 56,
    )

    draw.text((PAGE_W // 2, PAGE_H - 55), "TrendFlix Graduation Project", font=small_font, fill="#64748b", anchor="mm")
    return img


def main():
    OUT.parent.mkdir(parents=True, exist_ok=True)
    pages = [draw_page(i + 1, item, len(PAGES)) for i, item in enumerate(PAGES)]
    pages[0].save(OUT, "PDF", save_all=True, append_images=pages[1:], resolution=150.0)
    print(OUT)


if __name__ == "__main__":
    main()
