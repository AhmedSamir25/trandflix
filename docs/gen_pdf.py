#!/usr/bin/env python3
"""Generate TrendFlix code explanation PDF for new features (lists + watch later)."""

import arabic_reshaper
from bidi.algorithm import get_display
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import A4
from reportlab.lib import colors
from reportlab.lib.units import mm
import re

# ── Fonts ──────────────────────────────────────────────────────────────────────
ARABIC_FONT = '/usr/share/fonts/google-noto-vf/NotoSansArabic[wght].ttf'
pdfmetrics.registerFont(TTFont('NotoArabic', ARABIC_FONT))

# ── Colours (matching original PDF) ───────────────────────────────────────────
BG_WHITE       = colors.HexColor('#FFFFFF')
BG_DARK        = colors.HexColor('#1C1C2E')
BG_EXPLAIN     = colors.HexColor('#F5F5F5')
C_HEADER_GRAY  = colors.HexColor('#888888')
C_LINE_NUM     = colors.HexColor('#555577')
C_KEYWORD      = colors.HexColor('#E8C06E')   # yellow
C_STRING       = colors.HexColor('#5FCBCB')   # teal
C_COMMENT      = colors.HexColor('#6A6A8A')   # muted purple-gray
C_TYPE         = colors.HexColor('#7ECFFF')   # light blue
C_PLAIN        = colors.HexColor('#D4D4D4')   # light gray
C_RED_DOT      = colors.HexColor('#FF5F57')
C_YELLOW_DOT   = colors.HexColor('#FEBC2E')
C_GREEN_DOT    = colors.HexColor('#28C840')
C_SEPARATOR    = colors.HexColor('#DDDDDD')
C_FOOTER       = colors.HexColor('#AAAAAA')
C_EXPLAIN_HEAD = colors.HexColor('#222222')
C_EXPLAIN_BODY = colors.HexColor('#333333')
C_ACCENT       = colors.HexColor('#E53935')   # red accent for page nums

W, H = A4   # 595.27 x 841.89

MARGIN = 40
CODE_X = MARGIN
CODE_W = W - 2 * MARGIN
CODE_TOP = H - 115
CODE_BOTTOM = 260
CODE_H = CODE_TOP - CODE_BOTTOM

EXPLAIN_TOP = CODE_BOTTOM - 12
EXPLAIN_H = 170
EXPLAIN_BOTTOM = EXPLAIN_TOP - EXPLAIN_H

FOOTER_Y = 22

GO_KEYWORDS = {
    'func','if','else','return','var','type','struct','import','package',
    'for','range','switch','case','break','continue','nil','true','false',
    'make','new','go','defer','select','chan','map','interface','const',
    'error','string','uint','int','float64','bool','byte',
}


def ar(text):
    """Reshape + bidi Arabic text for correct rendering."""
    reshaped = arabic_reshaper.reshape(text)
    return get_display(reshaped)


def draw_page_chrome(c, page_num, total, arabic_title, total_pages_str=None):
    """Draw header separator and footer."""
    # Header left: page label
    c.setFont('Helvetica', 8)
    c.setFillColor(C_HEADER_GRAY)
    label = f'TrendFlix Code Explanation   {page_num:02d}/{total}'
    c.drawString(MARGIN, H - 42, label)

    # Header right: Arabic title
    c.setFont('NotoArabic', 18)
    c.setFillColor(C_EXPLAIN_HEAD)
    title_rtl = ar(arabic_title)
    c.drawRightString(W - MARGIN, H - 48, title_rtl)

    # Separator line
    c.setStrokeColor(C_SEPARATOR)
    c.setLineWidth(0.8)
    c.line(MARGIN, H - 60, W - MARGIN, H - 60)

    # Footer
    c.setFont('Helvetica', 8)
    c.setFillColor(C_FOOTER)
    footer = 'TrendFlix Graduation Project'
    c.drawCentredString(W / 2, FOOTER_Y, footer)


def draw_code_block(c, file_path, lines_data, box_top=None, box_bottom=None):
    """
    Draw a macOS-style dark terminal box with syntax-highlighted code.
    lines_data: list of (line_number_str, code_str)
    """
    bx = CODE_X
    by = box_bottom if box_bottom else CODE_BOTTOM
    bw = CODE_W
    bh = (box_top if box_top else CODE_TOP) - by

    # Background rounded rect
    c.setFillColor(BG_DARK)
    c.roundRect(bx, by, bw, bh, 8, fill=1, stroke=0)

    # Title bar row
    bar_h = 28
    bar_y = by + bh - bar_h
    c.setFillColor(colors.HexColor('#252540'))
    c.roundRect(bx, bar_y, bw, bar_h, 8, fill=1, stroke=0)
    # Cover lower-round of top rect
    c.setFillColor(colors.HexColor('#252540'))
    c.rect(bx, bar_y, bw, bar_h / 2, fill=1, stroke=0)

    # Traffic dots
    dot_y = bar_y + bar_h / 2
    dot_r = 5
    for col, dx in [(C_RED_DOT, 16), (C_YELLOW_DOT, 32), (C_GREEN_DOT, 48)]:
        c.setFillColor(col)
        c.circle(bx + dx, dot_y, dot_r, fill=1, stroke=0)

    # File path label
    c.setFont('Helvetica', 8)
    c.setFillColor(C_PLAIN)
    c.drawString(bx + 68, dot_y - 4, file_path)

    # Code lines
    code_area_top = bar_y - 6
    line_h = 13
    max_lines = int((code_area_top - by - 8) / line_h)
    display_lines = lines_data[:max_lines]

    for i, (lineno, code) in enumerate(display_lines):
        ly = code_area_top - (i + 1) * line_h
        if ly < by + 4:
            break

        # Line number
        c.setFont('Courier', 7.5)
        c.setFillColor(C_LINE_NUM)
        c.drawRightString(bx + 36, ly, str(lineno))

        # Tokenize + draw code
        draw_code_line(c, bx + 42, ly, code, bw - 50)


def draw_code_line(c, x, y, code, max_w):
    """Very simple tokenizer – enough for Go snippets."""
    tokens = tokenize_go(code)
    cursor = x
    for tok_type, tok_text in tokens:
        if cursor >= x + max_w:
            break
        if tok_type == 'keyword':
            col = C_KEYWORD
        elif tok_type == 'string':
            col = C_STRING
        elif tok_type == 'comment':
            col = C_COMMENT
        elif tok_type == 'number':
            col = colors.HexColor('#B5CEA8')
        else:
            col = C_PLAIN
        c.setFont('Courier', 7.5)
        c.setFillColor(col)
        c.drawString(cursor, y, tok_text)
        cursor += c.stringWidth(tok_text, 'Courier', 7.5)


def tokenize_go(line):
    """Yield (type, text) tokens for a line of Go code."""
    tokens = []
    i = 0
    s = line
    while i < len(s):
        # Comment
        if s[i:i+2] == '//':
            tokens.append(('comment', s[i:]))
            break
        # String literal "..."
        if s[i] == '"':
            j = i + 1
            while j < len(s) and s[j] != '"':
                if s[j] == '\\':
                    j += 1
                j += 1
            tokens.append(('string', s[i:j+1]))
            i = j + 1
            continue
        # Backtick string
        if s[i] == '`':
            j = i + 1
            while j < len(s) and s[j] != '`':
                j += 1
            tokens.append(('string', s[i:j+1]))
            i = j + 1
            continue
        # Word
        if s[i].isalpha() or s[i] == '_':
            j = i
            while j < len(s) and (s[j].isalnum() or s[j] == '_'):
                j += 1
            word = s[i:j]
            if word in GO_KEYWORDS:
                tokens.append(('keyword', word))
            else:
                tokens.append(('plain', word))
            i = j
            continue
        # Number
        if s[i].isdigit():
            j = i
            while j < len(s) and (s[j].isdigit() or s[j] in '.xXabcdefABCDEF'):
                j += 1
            tokens.append(('number', s[i:j]))
            i = j
            continue
        # Operator / punctuation
        tokens.append(('plain', s[i]))
        i += 1
    return tokens


def draw_explanation(c, title_suffix, body_ar):
    """Draw the Arabic explanation box."""
    bx = MARGIN
    by = EXPLAIN_BOTTOM
    bw = CODE_W
    bh = EXPLAIN_H

    c.setFillColor(BG_EXPLAIN)
    c.roundRect(bx, by, bw, bh, 6, fill=1, stroke=0)

    # Heading "شرح الجزء:"
    heading = ar('شرح الجزء:')
    c.setFont('NotoArabic', 13)
    c.setFillColor(C_EXPLAIN_HEAD)
    c.drawRightString(bx + bw - 16, by + bh - 26, heading)

    # Body text (wrapped Arabic)
    c.setFont('NotoArabic', 11)
    c.setFillColor(C_EXPLAIN_BODY)
    draw_arabic_paragraph(c, body_ar, bx + 16, by + bh - 42, bw - 32, 16)


def draw_arabic_paragraph(c, text, x, y, width, line_height):
    """Wrap and draw RTL Arabic paragraph."""
    words = text.split()
    lines = []
    current = []
    for w in words:
        test = ' '.join(current + [w])
        reshaped = arabic_reshaper.reshape(test)
        display = get_display(reshaped)
        if c.stringWidth(display, 'NotoArabic', 11) <= width:
            current.append(w)
        else:
            if current:
                lines.append(' '.join(current))
            current = [w]
    if current:
        lines.append(' '.join(current))

    cur_y = y
    for line in lines:
        if cur_y < EXPLAIN_BOTTOM + 4:
            break
        display_line = get_display(arabic_reshaper.reshape(line))
        c.drawRightString(x + width, cur_y, display_line)
        cur_y -= line_height


# ── Page data ─────────────────────────────────────────────────────────────────
PAGES = [
    {
        'title': 'نماذج القوائم والمشاهدة لاحقاً',
        'file': 'models/user_list_model.go  |  models/watch_later_model.go',
        'lines': [
            (1,  'package models'),
            (2,  ''),
            (3,  'import "time"'),
            (4,  ''),
            (5,  'type UserList struct {'),
            (6,  '    ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`'),
            (7,  '    UserID    uint      `gorm:"column:user_id;not null;index" json:"user_id"`'),
            (8,  '    Name      string    `gorm:"column:name;type:varchar(255);not null" json:"name"`'),
            (9,  '    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`'),
            (10, '    UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`'),
            (11, '}'),
            (12, ''),
            (13, 'func (UserList) TableName() string { return "user_lists" }'),
            (14, ''),
            (15, 'type UserListItem struct {'),
            (16, '    ID         uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`'),
            (17, '    UserListID uint      `gorm:"column:user_list_id;...uniqueIndex" json:"user_list_id"`'),
            (18, '    ItemID     uint      `gorm:"column:item_id;...uniqueIndex" json:"item_id"`'),
            (19, '    CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`'),
            (20, '}'),
            (21, ''),
            (22, 'type WatchLater struct {'),
            (23, '    ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`'),
            (24, '    UserID    uint      `gorm:"column:user_id;...uniqueIndex" json:"user_id"`'),
            (25, '    ItemID    uint      `gorm:"column:item_id;...uniqueIndex" json:"item_id"`'),
            (26, '    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`'),
            (27, '}'),
            (28, ''),
            (29, 'func (WatchLater) TableName() string { return "watch_later" }'),
        ],
        'explanation': 'يعرّف النظام ثلاثة نماذج جديدة: UserList تمثل قائمة مخصصة يصنعها المستخدم بنفسه مثل "أفلام العيد"، وUserListItem التي تربط كل عنصر بقائمته بعلاقة فريدة تمنع التكرار، وWatchLater التي تحفظ العناصر التي يريد المستخدم مشاهدتها لاحقاً. جميع النماذج تستخدم uniqueIndex لمنع إضافة نفس العنصر مرتين.',
    },
    {
        'title': 'مسارات القوائم والمشاهدة لاحقاً',
        'file': 'routers/lists_router.go  |  routers/watch_later_router.go',
        'lines': [
            (10, 'func RegisterListRoutes(app *fiber.App) {'),
            (11, '    lists := app.Group("/lists", middleware.Authenticate)'),
            (12, '    lists.Get("", listscontroller.GetLists)'),
            (13, '    lists.Post("", listscontroller.CreateList)'),
            (14, '    lists.Get("/:list_id", listscontroller.GetListItems)'),
            (15, '    lists.Post("/:list_id/items/:item_id", listscontroller.AddItemToList)'),
            (16, '    lists.Delete("/:list_id/items/:item_id", listscontroller.RemoveItemFromList)'),
            (17, '    lists.Delete("/:list_id", listscontroller.DeleteList)'),
            (18, '}'),
            (19, ''),
            (10, 'func RegisterWatchLaterRoutes(app *fiber.App) {'),
            (11, '    watchLater := app.Group("/watch-later", middleware.Authenticate)'),
            (12, '    watchLater.Get("", watchlatercontroller.GetWatchLater)'),
            (13, '    watchLater.Post("/:item_id", watchlatercontroller.AddWatchLater)'),
            (14, '    watchLater.Delete("/:item_id", watchlatercontroller.RemoveWatchLater)'),
            (15, '}'),
            (16, ''),
            (17, '// main.go additions:'),
            (18, 'routers.RegisterWatchLaterRoutes(app)'),
            (19, 'routers.RegisterListRoutes(app)'),
        ],
        'explanation': 'كل مسارات القوائم والمشاهدة لاحقاً محمية بـ middleware.Authenticate، أي يجب أن يكون المستخدم مسجلاً. القوائم تدعم إنشاء قائمة، جلبها، جلب محتواها، إضافة عنصر إليها وحذف عنصر أو القائمة كاملة. أما المشاهدة لاحقاً فهي أبسط: جلب، إضافة، وحذف عنصر واحد فقط.',
    },
    {
        'title': 'جلب قوائم المستخدم',
        'file': 'controller/lists_controller/lists_controller.go',
        'lines': [
            (16, 'func GetLists(c *fiber.Ctx) error {'),
            (17, '    context := fiber.Map{'),
            (18, '        "statusText": "Ok",'),
            (19, '        "msg":        "Lists fetched successfully",'),
            (20, '    }'),
            (21, ''),
            (22, '    if database.DbConn == nil {'),
            (23, '        log.Println("database connection is not initialized")'),
            (24, '        context["statusText"] = "bad"'),
            (25, '        context["msg"] = "Database error"'),
            (26, '        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (27, '    }'),
            (28, ''),
            (29, '    user, err := currentUserFromContext(c, context)'),
            (30, '    if err != nil {'),
            (31, '        return err'),
            (32, '    }'),
            (33, ''),
            (34, '    var lists []models.UserList'),
            (35, '    result := database.DbConn.'),
            (36, '        Where("user_id = ?", user.ID).'),
            (37, '        Order("created_at DESC").'),
            (38, '        Find(&lists)'),
            (39, '    if result.Error != nil {'),
            (40, '        log.Println("Error fetching lists:", result.Error)'),
            (41, '        context["statusText"] = "bad"'),
            (42, '        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (43, '    }'),
            (44, ''),
            (45, '    context["lists"] = lists'),
            (46, '    return c.Status(fiber.StatusOK).JSON(context)'),
            (47, '}'),
        ],
        'explanation': 'GetLists تجلب كل القوائم الخاصة بالمستخدم الحالي فقط. تستخدم currentUserFromContext لاستخراج المستخدم من Locals التي وضعها الـ middleware. ثم تستعلم قاعدة البيانات بشرط user_id مطابق وترتيب من الأحدث إلى الأقدم، وترجع القائمة بصيغة JSON.',
    },
    {
        'title': 'إنشاء قائمة جديدة',
        'file': 'controller/lists_controller/lists_controller.go',
        'lines': [
            (50, 'func CreateList(c *fiber.Ctx) error {'),
            (51, '    context := fiber.Map{'),
            (52, '        "statusText": "Ok",'),
            (53, '        "msg":        "List created successfully",'),
            (54, '    }'),
            (55, ''),
            (63, '    user, err := currentUserFromContext(c, context)'),
            (64, '    if err != nil { return err }'),
            (65, ''),
            (68, '    var req struct {'),
            (69, '        Name string `json:"name"`'),
            (70, '    }'),
            (71, '    if err := c.BodyParser(&req); err != nil {'),
            (72, '        context["msg"] = "Invalid request"'),
            (73, '        return c.Status(fiber.StatusBadRequest).JSON(context)'),
            (74, '    }'),
            (75, ''),
            (78, '    name := strings.TrimSpace(req.Name)'),
            (79, '    if name == "" {'),
            (80, '        context["msg"] = "List name is required"'),
            (81, '        return c.Status(fiber.StatusBadRequest).JSON(context)'),
            (82, '    }'),
            (83, ''),
            (85, '    list := models.UserList{UserID: user.ID, Name: name}'),
            (86, ''),
            (90, '    if err := database.DbConn.Create(&list).Error; err != nil {'),
            (91, '        context["msg"] = "Error saving list"'),
            (92, '        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (93, '    }'),
            (94, ''),
            (97, '    context["id"] = list.ID'),
            (98, '    context["list"] = list'),
            (99, '    return c.Status(fiber.StatusCreated).JSON(context)'),
            (100,'} '),
        ],
        'explanation': 'CreateList تستقبل name من جسم الطلب JSON وتنظفه بـ TrimSpace. إذا كان فارغاً ترجع BadRequest. بعد التحقق تبني نموذج UserList بربط user_id بالمستخدم الحالي تلقائياً ثم تحفظه. ترجع كود 201 Created مع بيانات القائمة الجديدة.',
    },
    {
        'title': 'حذف قائمة باستخدام Transaction',
        'file': 'controller/lists_controller/lists_controller.go',
        'lines': [
            (102,'func DeleteList(c *fiber.Ctx) error {'),
            (115,'    user, err := currentUserFromContext(c, context)'),
            (116,'    if err != nil { return err }'),
            (117,''),
            (120,'    listID, err := parseListID(c, context)'),
            (121,'    if err != nil { return err }'),
            (122,''),
            (125,'    tx := database.DbConn.Begin()'),
            (126,'    if tx.Error != nil {'),
            (127,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (128,'    }'),
            (129,''),
            (133,'    var list models.UserList'),
            (134,'    result := tx.Where("id = ? AND user_id = ?", listID, user.ID).First(&list)'),
            (135,'    if result.Error != nil {'),
            (136,'        tx.Rollback()'),
            (137,'        if errors.Is(result.Error, gorm.ErrRecordNotFound) {'),
            (138,'            context["msg"] = "List not found"'),
            (139,'            return c.Status(fiber.StatusNotFound).JSON(context)'),
            (140,'        }'),
            (141,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (142,'    }'),
            (143,''),
            (149,'    if err := tx.Where("user_list_id = ?", list.ID).Delete(&models.UserListItem{}).Error; err != nil {'),
            (150,'        tx.Rollback()'),
            (151,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (152,'    }'),
            (153,''),
            (157,'    if err := tx.Delete(&list).Error; err != nil {'),
            (158,'        tx.Rollback()'),
            (159,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (160,'    }'),
            (161,''),
            (165,'    if err := tx.Commit().Error; err != nil {'),
            (166,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (167,'    }'),
            (168,'    return c.Status(fiber.StatusOK).JSON(context)'),
            (169,'} '),
        ],
        'explanation': 'DeleteList تستخدم Transaction لأن الحذف يتم على جدولين: أولاً تحذف كل UserListItem المرتبطة بالقائمة، ثم تحذف القائمة نفسها. إذا فشلت أي خطوة يتم Rollback لضمان عدم بقاء بيانات يتيمة. أيضاً تتحقق أن القائمة تنتمي للمستخدم الحالي لمنع حذف قوائم الآخرين.',
    },
    {
        'title': 'جلب عناصر القائمة',
        'file': 'controller/lists_controller/lists_controller.go',
        'lines': [
            (175,'func GetListItems(c *fiber.Ctx) error {'),
            (187,'    user, err := currentUserFromContext(c, context)'),
            (188,'    if err != nil { return err }'),
            (189,''),
            (193,'    listID, err := parseListID(c, context)'),
            (194,'    if err != nil { return err }'),
            (195,''),
            (198,'    var list models.UserList'),
            (199,'    result := database.DbConn.Where("id = ? AND user_id = ?", listID, user.ID).First(&list)'),
            (200,'    if result.Error != nil {'),
            (201,'        if errors.Is(result.Error, gorm.ErrRecordNotFound) {'),
            (202,'            context["msg"] = "List not found"'),
            (203,'            return c.Status(fiber.StatusNotFound).JSON(context)'),
            (204,'        }'),
            (205,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (206,'    }'),
            (207,''),
            (213,'    var items []models.Item'),
            (214,'    result = database.DbConn.'),
            (215,'        Model(&models.Item{}).'),
            (216,'        Joins("JOIN user_list_items ON user_list_items.item_id = items.id").'),
            (217,'        Where("user_list_items.user_list_id = ?", list.ID).'),
            (218,'        Preload("Categories").'),
            (219,'        Order("user_list_items.created_at DESC").'),
            (220,'        Find(&items)'),
            (221,''),
            (228,'    context["list"] = list'),
            (229,'    context["items"] = items'),
            (230,'    return c.Status(fiber.StatusOK).JSON(context)'),
            (231,'} '),
        ],
        'explanation': 'GetListItems تتحقق أولاً أن القائمة المطلوبة تنتمي للمستخدم الحالي. بعد ذلك تجلب العناصر باستخدام JOIN بين جدول items وجدول user_list_items بشرط user_list_id، مع Preload للتصنيفات وترتيب من أحدث إضافة. ترجع بيانات القائمة والعناصر معاً في الاستجابة.',
    },
    {
        'title': 'إضافة عنصر إلى القائمة',
        'file': 'controller/lists_controller/lists_controller.go',
        'lines': [
            (233,'func AddItemToList(c *fiber.Ctx) error {'),
            (246,'    user, err := currentUserFromContext(c, context)'),
            (247,'    if err != nil { return err }'),
            (248,''),
            (251,'    listID, err := parseListID(c, context)'),
            (252,'    if err != nil { return err }'),
            (253,''),
            (256,'    var list models.UserList'),
            (257,'    result := database.DbConn.Where("id = ? AND user_id = ?", listID, user.ID).First(&list)'),
            (258,'    if result.Error != nil { /* 404 or 500 */ }'),
            (259,''),
            (271,'    itemID, err := parseItemID(c, context)'),
            (272,'    if err != nil { return err }'),
            (273,''),
            (276,'    if err := ensureItemExists(itemID, context, c); err != nil {'),
            (277,'        return err'),
            (278,'    }'),
            (279,''),
            (280,'    var item models.UserListItem'),
            (281,'    result = database.DbConn.Where("user_list_id = ? AND item_id = ?", list.ID, itemID).First(&item)'),
            (282,'    if result.Error == nil {'),
            (283,'        context["msg"] = "Item already in list"'),
            (284,'        return c.Status(fiber.StatusConflict).JSON(context)'),
            (285,'    }'),
            (286,''),
            (294,'    item = models.UserListItem{UserListID: list.ID, ItemID: itemID}'),
            (295,''),
            (299,'    if err := database.DbConn.Create(&item).Error; err != nil {'),
            (300,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (301,'    }'),
            (302,''),
            (306,'    context["id"] = item.ID'),
            (307,'    return c.Status(fiber.StatusCreated).JSON(context)'),
            (308,'} '),
        ],
        'explanation': 'AddItemToList تتحقق من ثلاثة أشياء قبل الحفظ: أن القائمة تنتمي للمستخدم، وأن العنصر موجود فعلاً في قاعدة البيانات عبر ensureItemExists، وأن العنصر لم يُضف مسبقاً لنفس القائمة. إذا وُجد مسبقاً ترجع Conflict. فقط عند نجاح كل الفحوصات تنشئ UserListItem وترجع 201 Created.',
    },
    {
        'title': 'حذف عنصر من القائمة',
        'file': 'controller/lists_controller/lists_controller.go',
        'lines': [
            (311,'func RemoveItemFromList(c *fiber.Ctx) error {'),
            (323,'    user, err := currentUserFromContext(c, context)'),
            (324,'    if err != nil { return err }'),
            (325,''),
            (328,'    listID, err := parseListID(c, context)'),
            (329,'    if err != nil { return err }'),
            (330,''),
            (333,'    var list models.UserList'),
            (334,'    result := database.DbConn.Where("id = ? AND user_id = ?", listID, user.ID).First(&list)'),
            (335,'    if result.Error != nil {'),
            (336,'        if errors.Is(result.Error, gorm.ErrRecordNotFound) {'),
            (337,'            context["msg"] = "List not found"'),
            (338,'            return c.Status(fiber.StatusNotFound).JSON(context)'),
            (339,'        }'),
            (340,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (341,'    }'),
            (342,''),
            (348,'    itemID, err := parseItemID(c, context)'),
            (349,'    if err != nil { return err }'),
            (350,''),
            (353,'    var item models.UserListItem'),
            (354,'    result = database.DbConn.Where("user_list_id = ? AND item_id = ?", list.ID, itemID).First(&item)'),
            (355,'    if result.Error != nil {'),
            (356,'        if errors.Is(result.Error, gorm.ErrRecordNotFound) {'),
            (357,'            context["msg"] = "Item not found in list"'),
            (358,'            return c.Status(fiber.StatusNotFound).JSON(context)'),
            (359,'        }'),
            (360,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (361,'    }'),
            (362,''),
            (369,'    if err := database.DbConn.Delete(&item).Error; err != nil {'),
            (370,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (371,'    }'),
            (372,'    return c.Status(fiber.StatusOK).JSON(context)'),
            (373,'} '),
        ],
        'explanation': 'RemoveItemFromList تتحقق أولاً أن القائمة للمستخدم الحالي، ثم تبحث عن UserListItem المطلوبة بشرط user_list_id و item_id معاً. إذا لم توجد ترجع 404. عند وجودها تحذفها مباشرة بدون transaction لأن الحذف يطال جدولاً واحداً فقط.',
    },
    {
        'title': 'جلب قائمة المشاهدة لاحقاً',
        'file': 'controller/watch_later_controller/watch_later_controller.go',
        'lines': [
            (16, 'func GetWatchLater(c *fiber.Ctx) error {'),
            (17, '    context := fiber.Map{'),
            (18, '        "statusText": "Ok",'),
            (19, '        "msg":        "Watch later items fetched successfully",'),
            (20, '    }'),
            (21, ''),
            (22, '    if database.DbConn == nil {'),
            (23, '        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (24, '    }'),
            (25, ''),
            (29, '    user, err := currentUserFromContext(c, context)'),
            (30, '    if err != nil { return err }'),
            (31, ''),
            (34, '    var items []models.Item'),
            (35, '    result := database.DbConn.'),
            (36, '        Model(&models.Item{}).'),
            (37, '        Joins("JOIN watch_later ON watch_later.item_id = items.id").'),
            (38, '        Where("watch_later.user_id = ?", user.ID).'),
            (39, '        Preload("Categories").'),
            (40, '        Order("watch_later.created_at DESC").'),
            (41, '        Find(&items)'),
            (42, '    if result.Error != nil {'),
            (43, '        log.Println("Error fetching watch later:", result.Error)'),
            (44, '        context["statusText"] = "bad"'),
            (45, '        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (46, '    }'),
            (47, ''),
            (49, '    context["items"] = items'),
            (50, '    return c.Status(fiber.StatusOK).JSON(context)'),
            (51, '} '),
        ],
        'explanation': 'GetWatchLater تجلب العناصر التي أضافها المستخدم للمشاهدة لاحقاً. تستخدم JOIN مباشر بين جدول items وجدول watch_later بشرط user_id للمستخدم الحالي. تحمّل التصنيفات بـ Preload وترتب النتائج من الأحدث إضافةً. هذه الدالة مشابهة لـ GetFavorites لكن تستخدم جدول watch_later بدل favorites.',
    },
    {
        'title': 'إضافة عنصر للمشاهدة لاحقاً',
        'file': 'controller/watch_later_controller/watch_later_controller.go',
        'lines': [
            (53, 'func AddWatchLater(c *fiber.Ctx) error {'),
            (54, '    context := fiber.Map{'),
            (55, '        "statusText": "Ok",'),
            (56, '        "msg":        "Item added to watch later successfully",'),
            (57, '    }'),
            (58, ''),
            (66, '    user, err := currentUserFromContext(c, context)'),
            (67, '    if err != nil { return err }'),
            (68, ''),
            (71, '    itemID, err := parseItemID(c, context)'),
            (72, '    if err != nil { return err }'),
            (73, ''),
            (76, '    if err := ensureItemExists(itemID, context, c); err != nil {'),
            (77, '        return err'),
            (78, '    }'),
            (79, ''),
            (80, '    var wl models.WatchLater'),
            (81, '    result := database.DbConn.Where("user_id = ? AND item_id = ?", user.ID, itemID).First(&wl)'),
            (82, '    if result.Error == nil {'),
            (83, '        context["msg"] = "Item already in watch later"'),
            (84, '        return c.Status(fiber.StatusConflict).JSON(context)'),
            (85, '    }'),
            (86, ''),
            (94, '    wl = models.WatchLater{UserID: user.ID, ItemID: itemID}'),
            (95, ''),
            (99, '    if err := database.DbConn.Create(&wl).Error; err != nil {'),
            (100,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (101,'    }'),
            (102,''),
            (106,'    context["id"] = wl.ID'),
            (107,'    context["watchLater"] = wl'),
            (108,'    return c.Status(fiber.StatusCreated).JSON(context)'),
            (109,'} '),
        ],
        'explanation': 'AddWatchLater تضيف عنصراً لقائمة المشاهدة لاحقاً للمستخدم الحالي. تتحقق أولاً أن العنصر موجود في قاعدة البيانات عبر ensureItemExists، ثم تبحث إن كان قد أُضيف مسبقاً فإن وُجد ترجع Conflict. عند نجاح كل الفحوصات تنشئ سجل WatchLater بـ user_id و item_id وترجع 201 Created.',
    },
    {
        'title': 'حذف عنصر من المشاهدة لاحقاً',
        'file': 'controller/watch_later_controller/watch_later_controller.go',
        'lines': [
            (111,'func RemoveWatchLater(c *fiber.Ctx) error {'),
            (112,'    context := fiber.Map{'),
            (113,'        "statusText": "Ok",'),
            (114,'        "msg":        "Item removed from watch later successfully",'),
            (115,'    }'),
            (116,''),
            (122,'    user, err := currentUserFromContext(c, context)'),
            (123,'    if err != nil { return err }'),
            (124,''),
            (128,'    itemID, err := parseItemID(c, context)'),
            (129,'    if err != nil { return err }'),
            (130,''),
            (133,'    var wl models.WatchLater'),
            (134,'    result := database.DbConn.Where("user_id = ? AND item_id = ?", user.ID, itemID).First(&wl)'),
            (135,'    if result.Error != nil {'),
            (136,'        if errors.Is(result.Error, gorm.ErrRecordNotFound) {'),
            (137,'            context["statusText"] = "bad"'),
            (138,'            context["msg"] = "Watch later item not found"'),
            (139,'            return c.Status(fiber.StatusNotFound).JSON(context)'),
            (140,'        }'),
            (141,'        log.Println("Error querying watch later:", result.Error)'),
            (142,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (143,'    }'),
            (144,''),
            (149,'    if err := database.DbConn.Delete(&wl).Error; err != nil {'),
            (150,'        log.Println("Error deleting watch later:", err)'),
            (151,'        context["statusText"] = "bad"'),
            (152,'        context["msg"] = "Error deleting watch later"'),
            (153,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (154,'    }'),
            (155,''),
            (156,'    return c.Status(fiber.StatusOK).JSON(context)'),
            (157,'} '),
        ],
        'explanation': 'RemoveWatchLater تحذف عنصراً من قائمة المشاهدة لاحقاً للمستخدم الحالي. تبحث في جدول watch_later بشرط user_id و item_id معاً؛ إذا لم يوجد ترجع 404. إذا وُجد تحذفه مباشرة. لا تحتاج Transaction لأن الحذف على سجل واحد في جدول واحد.',
    },
    {
        'title': 'دوال مساعدة مشتركة',
        'file': 'controller/lists_controller  |  controller/watch_later_controller',
        'lines': [
            (379,'func currentUserFromContext(c *fiber.Ctx, context fiber.Map) (models.User, error) {'),
            (380,'    userValue := c.Locals("currentUser")'),
            (381,'    user, ok := userValue.(models.User)'),
            (382,'    if !ok {'),
            (383,'        context["statusText"] = "bad"'),
            (384,'        context["msg"] = "Unauthorized"'),
            (385,'        return models.User{}, c.Status(fiber.StatusUnauthorized).JSON(context)'),
            (386,'    }'),
            (387,'    return user, nil'),
            (388,'}'),
            (389,''),
            (391,'func parseListID(c *fiber.Ctx, context fiber.Map) (uint, error) {'),
            (392,'    listID, err := strconv.ParseUint(strings.TrimSpace(c.Params("list_id")), 10, 64)'),
            (393,'    if err != nil || listID == 0 {'),
            (394,'        context["msg"] = "Invalid list id"'),
            (395,'        return 0, c.Status(fiber.StatusBadRequest).JSON(context)'),
            (396,'    }'),
            (397,'    return uint(listID), nil'),
            (398,'}'),
            (399,''),
            (413,'func ensureItemExists(itemID uint, context fiber.Map, c *fiber.Ctx) error {'),
            (414,'    var item models.Item'),
            (415,'    result := database.DbConn.Select("id").First(&item, itemID)'),
            (416,'    if result.Error != nil {'),
            (417,'        if errors.Is(result.Error, gorm.ErrRecordNotFound) {'),
            (418,'            context["msg"] = "Item not found"'),
            (419,'            return c.Status(fiber.StatusNotFound).JSON(context)'),
            (420,'        }'),
            (421,'        return c.Status(fiber.StatusInternalServerError).JSON(context)'),
            (422,'    }'),
            (423,'    return nil'),
            (424,'}'),
        ],
        'explanation': 'الدوال المساعدة مشتركة بين كلا الـ controllers. currentUserFromContext تستخرج المستخدم من Locals وإذا لم توجد تعيد Unauthorized. parseListID و parseItemID تحوّل المعامل من نص إلى uint مع التحقق من صحته. ensureItemExists تتأكد من وجود العنصر في قاعدة البيانات قبل أي عملية تضمن سلامة البيانات.',
    },
]

TOTAL = len(PAGES)


def generate_pdf(output_path):
    c = canvas.Canvas(output_path, pagesize=A4)
    c.setTitle('TrendFlix - Lists & Watch Later Feature Explanation')

    for i, page in enumerate(PAGES, start=1):
        c.setFillColor(BG_WHITE)
        c.rect(0, 0, W, H, fill=1, stroke=0)

        draw_page_chrome(c, i, TOTAL, page['title'])
        draw_code_block(c, page['file'], [(ln, code) for ln, code in page['lines']])
        draw_explanation(c, '', page['explanation'])

        if i < TOTAL:
            c.showPage()

    c.save()
    print(f'PDF saved to {output_path}')


if __name__ == '__main__':
    import os
    out = os.path.join(os.path.dirname(__file__), 'trendflix_lists_watchlater_ar.pdf')
    generate_pdf(out)
