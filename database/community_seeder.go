package database

import (
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"trendflix/models"
)

const defaultSeedUserPassword = "password123"

type seedUser struct {
	Name   string
	Email  string
	Avatar string
	Role   string
}

type seedCommunity struct {
	Name         string
	Slug         string
	Description  string
	Rules        string
	CategoryType string
	CreatorEmail string
	IsPrivate    bool
	MemberEmails []string
}

type seedPost struct {
	CommunitySlug string
	AuthorEmail   string
	Title         string
	Body          string
	PostType      string
	IsSpoiler     bool
	IsPinned      bool
}

var defaultSeedUsers = []seedUser{
	{Name: "Ahmed Hassan", Email: "ahmed@trendflix.local", Avatar: "https://i.pravatar.cc/150?img=12", Role: "user"},
	{Name: "Sara Khalid", Email: "sara@trendflix.local", Avatar: "https://i.pravatar.cc/150?img=45", Role: "user"},
	{Name: "Omar Farouk", Email: "omar@trendflix.local", Avatar: "https://i.pravatar.cc/150?img=33", Role: "user"},
	{Name: "Layla Ibrahim", Email: "layla@trendflix.local", Avatar: "https://i.pravatar.cc/150?img=47", Role: "user"},
	{Name: "Yusuf Nasser", Email: "yusuf@trendflix.local", Avatar: "https://i.pravatar.cc/150?img=51", Role: "user"},
	{Name: "Fatima Zahra", Email: "fatima@trendflix.local", Avatar: "https://i.pravatar.cc/150?img=49", Role: "user"},
	{Name: "John Carter", Email: "john@trendflix.local", Avatar: "https://i.pravatar.cc/150?img=15", Role: "user"},
	{Name: "Mei Lin", Email: "mei@trendflix.local", Avatar: "https://i.pravatar.cc/150?img=20", Role: "user"},
}

var defaultSeedCommunities = []seedCommunity{
	{
		Name:         "Sci-Fi Lovers",
		Slug:         "sci-fi-lovers",
		CategoryType: models.CategoryTypeMixed,
		CreatorEmail: "ahmed@trendflix.local",
		Description:  "A community for everyone who loves science fiction — movies, series, books, and games. Discuss time travel, space opera, AI, and the futures we dream about.\n\nمجتمع لكل محبي الخيال العلمي من أفلام ومسلسلات وكتب وألعاب. نتناقش حول السفر عبر الزمن، الملاحم الفضائية، الذكاء الاصطناعي، والمستقبلات التي نحلم بها.",
		Rules:        "1. Be respectful. / كن محترماً.\n2. No spoilers without tags. / ممنوع الحرق بدون تحذير.\n3. Stay on topic. / التزم بالموضوع.",
		IsPrivate:    false,
		MemberEmails: []string{
			"ahmed@trendflix.local",
			"sara@trendflix.local",
			"omar@trendflix.local",
			"john@trendflix.local",
			"mei@trendflix.local",
		},
	},
	{
		Name:         "Cinephiles",
		Slug:         "cinephiles",
		CategoryType: models.CategoryTypeMovies,
		CreatorEmail: "sara@trendflix.local",
		Description:  "For the people who live and breathe cinema. From classics to the latest releases, share reviews, analyses, and hidden gems.\n\nلعشاق السينما الحقيقيين. من الكلاسيكيات إلى أحدث الإصدارات، شارك مراجعاتك وتحليلاتك والأفلام المخفية.",
		Rules:        "1. Use the review post type for reviews.\n2. Tag spoilers.\n3. Respect different tastes.",
		IsPrivate:    false,
		MemberEmails: []string{
			"sara@trendflix.local",
			"layla@trendflix.local",
			"ahmed@trendflix.local",
			"yusuf@trendflix.local",
			"john@trendflix.local",
		},
	},
	{
		Name:         "Gamers Hub",
		Slug:         "gamers-hub",
		CategoryType: models.CategoryTypeGames,
		CreatorEmail: "omar@trendflix.local",
		Description:  "Everything gaming. New releases, indie discoveries, recommendations, and what to play next.\n\nكل ما يخص الألعاب. الإصدارات الجديدة، الاكتشافات المستقلة، التوصيات، وماذا تلعب بعد ذلك.",
		Rules:        "1. No piracy links.\n2. Mark spoilers.\n3. Keep platform wars civil.",
		IsPrivate:    false,
		MemberEmails: []string{
			"omar@trendflix.local",
			"yusuf@trendflix.local",
			"ahmed@trendflix.local",
			"mei@trendflix.local",
		},
	},
	{
		Name:         "Book Club",
		Slug:         "book-club",
		CategoryType: models.CategoryTypeBooks,
		CreatorEmail: "layla@trendflix.local",
		Description:  "A quiet corner for readers. Monthly reads, deep discussions, and recommendations across every genre.\n\nركن هادئ للقرّاء. قراءات شهرية، نقاشات عميقة، وتوصيات عبر كل الأنواع.",
		Rules:        "1. Spoilers must be tagged.\n2. Recommend before you critique.\n3. Welcome all genres.",
		IsPrivate:    false,
		MemberEmails: []string{
			"layla@trendflix.local",
			"fatima@trendflix.local",
			"sara@trendflix.local",
		},
	},
	{
		Name:         "Series Watchers",
		Slug:         "series-watchers",
		CategoryType: models.CategoryTypeSeries,
		CreatorEmail: "fatima@trendflix.local",
		Description:  "Binge-worthy conversations about TV shows and limited series. Episode threads, theories, and what deserves your weekend.\n\nنقاشات لا تُقاوَم حول المسلسلات. مواضيع الحلقات، النظريات، وما يستحق عطلة نهاية أسبوعك.",
		Rules:        "1. Tag episode spoilers.\n2. Be kind about hot takes.\n3. Share where to watch legally.",
		IsPrivate:    false,
		MemberEmails: []string{
			"fatima@trendflix.local",
			"ahmed@trendflix.local",
			"layla@trendflix.local",
			"mei@trendflix.local",
			"john@trendflix.local",
		},
	},
}

var defaultSeedPosts = []seedPost{
	{
		CommunitySlug: "sci-fi-lovers",
		AuthorEmail:   "ahmed@trendflix.local",
		Title:         "Is Interstellar the best modern sci-fi movie?",
		Body:          "I rewatched Interstellar this weekend and it still hits just as hard. The blend of real physics with emotion is rare. What do you think — is it the best modern sci-fi film, or is there one you'd rank higher?\n\nأعدت مشاهدة Interstellar هذا الأسبوع ولا يزال يؤثر فيّ بنفس القوة. المزج بين الفيزياء الحقيقية والمشاعر أمر نادر. ما رأيكم — هل هو أفضل فيلم خيال علمي حديث، أم هناك ما هو أفضل في نظركم؟",
		PostType:      models.PostTypeDiscussion,
		IsPinned:      true,
	},
	{
		CommunitySlug: "sci-fi-lovers",
		AuthorEmail:   "john@trendflix.local",
		Title:         "Dune vs. The Matrix — which world would you live in?",
		Body:          "Two huge sci-fi universes, very different vibes. Arrakis with its politics and prophecy, or the Matrix with its reality-bending rebellion? Pick one and tell me why.\n\nعالمان ضخمان من الخيال العلمي وبأجواء مختلفة جداً. كوكب أراكسيس بسياساته ونبوءاته، أم الماتريكس بتمرده الذي يثني الواقع؟ اختر واحداً وأخبرني لماذا.",
		PostType:      models.PostTypeDiscussion,
	},
	{
		CommunitySlug: "sci-fi-lovers",
		AuthorEmail:   "mei@trendflix.local",
		Title:         "Recommendations for someone who loved Dark (the series)",
		Body:          "Dark messed with my head in the best way — time loops, family secrets, perfect pacing. Looking for movies or shows that scratch the same itch. Any suggestions?\n\nمسلسل Dark أربك عقلي بشكل رائع — حلقات زمنية، أسرار عائلية، وإيقاع مثالي. أبحث عن أفلام أو مسلسلات تعطيني الشعور نفسه. أي اقتراحات؟",
		PostType:      models.PostTypeRecommendationRequest,
	},
	{
		CommunitySlug: "cinephiles",
		AuthorEmail:   "sara@trendflix.local",
		Title:         "The cinematography of Parasite — a masterclass",
		Body:          "Bong Joon-ho uses vertical space to tell the whole class story: the basement, the stairs, the rich house on the hill. Every frame is intentional. Let's talk about the shots that stayed with you.\n\nبونغ جون هو يستخدم المساحة العمودية ليحكي قصة الطبقات بأكملها: القبو، الدرج، بيت الأثرياء على التلة. كل لقطة مقصودة. لنتحدث عن اللقطات التي بقيت في ذاكرتكم.",
		PostType:      models.PostTypeReview,
		IsPinned:      true,
	},
	{
		CommunitySlug: "cinephiles",
		AuthorEmail:   "yusuf@trendflix.local",
		Title:         "Underrated crime thrillers you never hear about",
		Body:          "Everyone knows The Godfather and Pulp Fiction. What are the crime thrillers flying under the radar that more people should watch? Drop your list.\n\nالجميع يعرف The Godfather و Pulp Fiction. ما هي أفلام الجريمة والإثارة التي تمر دون ضجيج ويجب أن يشاهدها المزيد من الناس؟ شارك قائمتك.",
		PostType:      models.PostTypeDiscussion,
	},
	{
		CommunitySlug: "cinephiles",
		AuthorEmail:   "john@trendflix.local",
		Title:         "Is the 3-hour runtime in Oppenheimer justified?",
		Body:          "I'll be honest — I was never bored, but I also felt the length. Curious how others felt about the pacing. Did the runtime earn itself?\n\nبصراحة — لم أشعر بالملل أبداً، لكنني شعرت بالطول أيضاً. فضولي كيف شعر الباقون بالإيقاع. هل استحق الفيلم مدته الزمنية؟",
		PostType:      models.PostTypeDiscussion,
	},
	{
		CommunitySlug: "gamers-hub",
		AuthorEmail:   "omar@trendflix.local",
		Title:         "Elden Ring vs. Baldur's Gate 3 — game of the generation?",
		Body:          "Two giants from recent years. One perfected open-world exploration and challenge, the other perfected choice and storytelling. Which one is your game of the generation?\n\nعمالقة يافعان من السنوات الأخيرة. أحدهما أتقن الاستكشاف والتحدي في العالم المفتوح، والآخر أتقن الخيارات والسرد. أيهما لعبة الجيل بالنسبة لك؟",
		PostType:      models.PostTypeDiscussion,
		IsPinned:      true,
	},
	{
		CommunitySlug: "gamers-hub",
		AuthorEmail:   "yusuf@trendflix.local",
		Title:         "Cozy games to relax with after a long day",
		Body:          "Stardew Valley and Vampire Survivors have been my comfort food lately. What cozy or low-stress games do you unwind with?\n\nأصبحت Stardew Valley و Vampire Survivors طعامي المريح مؤخراً. ما هي الألعاب المريحة أو قليلة التوتر التي تسترخي بها؟",
		PostType:      models.PostTypeDiscussion,
	},
	{
		CommunitySlug: "gamers-hub",
		AuthorEmail:   "mei@trendflix.local",
		Title:         "What should I play next on a low-end PC?",
		Body:          "My laptop isn't powerful, but I want something engaging. I like RPGs and metroidvanias. Any recommendations that run well on modest hardware?\n\nلابتوبي ليس قوياً، لكنني أريد شيئاً مشوقاً. أحب ألعاب الـ RPG والميتروفانيا. أي توصيات تعمل بشكل جيد على جهاز متواضع؟",
		PostType:      models.PostTypeRecommendationRequest,
	},
	{
		CommunitySlug: "book-club",
		AuthorEmail:   "layla@trendflix.local",
		Title:         "Our November read: The Kite Runner — discussion thread",
		Body:          "Let's keep spoilers in this thread only. What did everyone think of the ending? I found the themes of guilt and redemption incredibly powerful.\n\nلنحتفظ بالحرق في هذا الموضوع فقط. ما رأي الجميع في النهاية؟ وجدت موضوعي الذنب والفداء مؤثرين للغاية.",
		PostType:      models.PostTypeDiscussion,
		IsPinned:      true,
	},
	{
		CommunitySlug: "book-club",
		AuthorEmail:   "fatima@trendflix.local",
		Title:         "Books that changed how you think",
		Body:          "Sapiens completely shifted how I see history and society. What's a book that genuinely changed your perspective on something?\n\nكتاب Sapiens غيّر طريقة رؤيتي للتاريخ والمجتمع تماماً. ما هو الكتاب الذي غيّر نظرتك حقاً تجاه شيء ما؟",
		PostType:      models.PostTypeDiscussion,
	},
	{
		CommunitySlug: "book-club",
		AuthorEmail:   "sara@trendflix.local",
		Title:         "Arabic novels everyone should read at least once",
		Body:          "Looking to read more Arabic literature. Beyond Naguib Mahfouz, which Arabic novels are essential? Would love your favorites.\n\nأبحث عن قراءة المزيد من الأدب العربي. غير نجيب محفوظ، ما هي الروايات العربية الأساسية؟ يسعدني معرفة المفضلة لديكم.",
		PostType:      models.PostTypeRecommendationRequest,
	},
	{
		CommunitySlug: "series-watchers",
		AuthorEmail:   "fatima@trendflix.local",
		Title:         "Breaking Bad vs. Better Call Saul — which is better?",
		Body:          "Hot take: I think Better Call Saul is the better-written show, even though Breaking Bad is more iconic. The character work in BCS is unreal. Fight me politely.\n\nرأي جريء: أعتقد أن Better Call Saul مكتوب بشكل أفضل، رغم أن Breaking Bad أكثر أيقونية. البناء الدرامي للشخصيات في BCS خيالي. خالفوني بأدب.",
		PostType:      models.PostTypeDiscussion,
		IsPinned:      true,
	},
	{
		CommunitySlug: "series-watchers",
		AuthorEmail:   "john@trendflix.local",
		Title:         "The Last of Us — best video game adaptation ever?",
		Body:          "They actually did it justice. The casting, the pacing, the respect for the source material. Is this now the bar for game adaptations?\n\nلقد أنصفوه فعلاً. اختيار الممثلين، الإيقاع، والاحترام للأصل. هل أصبح هذا هو المعيار لتكييفات الألعاب؟",
		PostType:      models.PostTypeDiscussion,
	},
	{
		CommunitySlug: "series-watchers",
		AuthorEmail:   "mei@trendflix.local",
		Title:         "Short series (one season) that are worth your time",
		Body:          "Not everything needs 8 seasons. What limited series told a complete, excellent story in one go? I loved Chernobyl and Queen's Gambit.\n\nليس كل شيء يحتاج 8 مواسم. أي مسلسل محدود روى قصة كاملة وممتازة دفعة واحدة؟ أحببت Chernobyl و Queen's Gambit.",
		PostType:      models.PostTypeDiscussion,
	},
}

func SeedCommunityContent() {
	if DbConn == nil {
		panic("database is not connected")
	}

	tx := DbConn.Begin()
	if tx.Error != nil {
		panic(fmt.Sprintf("community seed transaction failed: %v", tx.Error))
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultSeedUserPassword), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		panic(fmt.Sprintf("community seed password hash failed: %v", err))
	}

	userMap, usersCreated, err := ensureSeedUsers(tx, hashedPassword)
	if err != nil {
		tx.Rollback()
		panic(fmt.Sprintf("community seed users failed: %v", err))
	}

	communityMap, communitiesCreated, membersCreated, err := ensureSeedCommunities(tx, userMap)
	if err != nil {
		tx.Rollback()
		panic(fmt.Sprintf("community seed communities failed: %v", err))
	}

	postsCreated, err := ensureSeedPosts(tx, userMap, communityMap)
	if err != nil {
		tx.Rollback()
		panic(fmt.Sprintf("community seed posts failed: %v", err))
	}

	if err := refreshSeedCounters(tx, communityMap); err != nil {
		tx.Rollback()
		panic(fmt.Sprintf("community seed counters failed: %v", err))
	}

	if err := tx.Commit().Error; err != nil {
		panic(fmt.Sprintf("community seed commit failed: %v", err))
	}

	log.Printf(
		"community seed: created %d users, %d communities, %d members, %d posts (login: any seeded user with password %q)",
		usersCreated,
		communitiesCreated,
		membersCreated,
		postsCreated,
		defaultSeedUserPassword,
	)
}

func ensureSeedUsers(tx *gorm.DB, hashedPassword []byte) (map[string]models.User, int, error) {
	userMap := make(map[string]models.User, len(defaultSeedUsers))
	created := 0

	for _, entry := range defaultSeedUsers {
		var existing models.User
		result := tx.Where("email = ?", entry.Email).First(&existing)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, 0, result.Error
		}

		if result.RowsAffected > 0 {
			userMap[entry.Email] = existing
			continue
		}

		role := entry.Role
		if role == "" {
			role = "user"
		}

		user := models.User{
			Name:     entry.Name,
			Email:    entry.Email,
			Password: string(hashedPassword),
			Avatar:   entry.Avatar,
			Role:     role,
		}

		if err := tx.Create(&user).Error; err != nil {
			return nil, 0, err
		}

		userMap[entry.Email] = user
		created++
	}

	return userMap, created, nil
}

func ensureSeedCommunities(tx *gorm.DB, userMap map[string]models.User) (map[string]models.Community, int, int, error) {
	communityMap := make(map[string]models.Community, len(defaultSeedCommunities))
	communitiesCreated := 0
	membersCreated := 0

	for _, entry := range defaultSeedCommunities {
		creator, ok := userMap[entry.CreatorEmail]
		if !ok {
			return nil, 0, 0, fmt.Errorf("missing creator user %q for community %q", entry.CreatorEmail, entry.Slug)
		}

		var existing models.Community
		result := tx.Where("slug = ?", entry.Slug).First(&existing)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, 0, 0, result.Error
		}

		var community models.Community
		if result.RowsAffected > 0 {
			community = existing
		} else {
			categoryType := entry.CategoryType
			if categoryType == "" {
				categoryType = models.CategoryTypeMixed
			}

			community = models.Community{
				Name:         entry.Name,
				Slug:         entry.Slug,
				Description:  entry.Description,
				Rules:        entry.Rules,
				CategoryType: categoryType,
				CreatedBy:    creator.ID,
				IsPrivate:    entry.IsPrivate,
				Status:       models.CommunityStatusApproved,
			}

			if err := tx.Create(&community).Error; err != nil {
				return nil, 0, 0, err
			}
			communitiesCreated++
		}

		communityMap[entry.Slug] = community

		memberEmails := append([]string{entry.CreatorEmail}, entry.MemberEmails...)
		seen := make(map[string]bool, len(memberEmails))
		for _, email := range memberEmails {
			if seen[email] {
				continue
			}
			seen[email] = true

			member, ok := userMap[email]
			if !ok {
				return nil, 0, 0, fmt.Errorf("missing member user %q for community %q", email, entry.Slug)
			}

			created, err := ensureSeedMember(tx, community.ID, member.ID, email == entry.CreatorEmail)
			if err != nil {
				return nil, 0, 0, err
			}
			if created {
				membersCreated++
			}
		}
	}

	return communityMap, communitiesCreated, membersCreated, nil
}

func ensureSeedMember(tx *gorm.DB, communityID, userID uint, isCreator bool) (bool, error) {
	var existing models.CommunityMember
	result := tx.Where("community_id = ? AND user_id = ?", communityID, userID).First(&existing)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return false, nil
	}

	role := models.MemberRoleMember
	if isCreator {
		role = models.MemberRoleAdmin
	}

	member := models.CommunityMember{
		CommunityID: communityID,
		UserID:      userID,
		Role:        role,
		Status:      models.MemberStatusActive,
	}

	if err := tx.Create(&member).Error; err != nil {
		return false, err
	}

	return true, nil
}

func ensureSeedPosts(tx *gorm.DB, userMap map[string]models.User, communityMap map[string]models.Community) (int, error) {
	created := 0

	for _, entry := range defaultSeedPosts {
		community, ok := communityMap[entry.CommunitySlug]
		if !ok {
			return 0, fmt.Errorf("missing community slug %q for post %q", entry.CommunitySlug, entry.Title)
		}

		author, ok := userMap[entry.AuthorEmail]
		if !ok {
			return 0, fmt.Errorf("missing author user %q for post %q", entry.AuthorEmail, entry.Title)
		}

		var existing models.Post
		result := tx.Where("community_id = ? AND user_id = ? AND title = ?", community.ID, author.ID, entry.Title).First(&existing)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, result.Error
		}
		if result.RowsAffected > 0 {
			continue
		}

		postType := entry.PostType
		if postType == "" {
			postType = models.PostTypeDiscussion
		}

		post := models.Post{
			CommunityID: community.ID,
			UserID:      author.ID,
			Title:       entry.Title,
			Body:        entry.Body,
			PostType:    postType,
			IsSpoiler:   entry.IsSpoiler,
			IsPinned:    entry.IsPinned,
			Status:      models.PostStatusPublished,
		}

		if err := tx.Create(&post).Error; err != nil {
			return 0, err
		}
		created++
	}

	return created, nil
}

func refreshSeedCounters(tx *gorm.DB, communityMap map[string]models.Community) error {
	for _, community := range communityMap {
		var membersCount int64
		if err := tx.Model(&models.CommunityMember{}).
			Where("community_id = ? AND status = ?", community.ID, models.MemberStatusActive).
			Count(&membersCount).Error; err != nil {
			return err
		}

		var postsCount int64
		if err := tx.Model(&models.Post{}).
			Where("community_id = ? AND status = ?", community.ID, models.PostStatusPublished).
			Count(&postsCount).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Community{}).Where("id = ?", community.ID).Updates(map[string]interface{}{
			"members_count": membersCount,
			"posts_count":   postsCount,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}
