package database

import (
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"trendflix/models"
)

type seedItem struct {
	Title         string
	Description   string
	Type          string
	CoverImage    string
	ContentLink   string
	ReleaseDate   string
	Author        string
	Director      string
	Developer     string
	Duration      uint
	PagesCount    uint
	Platform      string
	Rating        float64
	CategorySlugs []string
}

var defaultItems = []seedItem{
	{
		Title:         "To Kill a Mockingbird",
		Description:   "Harper Lee's classic novel follows Scout Finch as she confronts racism and injustice in the American South.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780061120084-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=To+Kill+a+Mockingbird+Harper+Lee+book",
		ReleaseDate:   "1960-07-11",
		Author:        "Harper Lee",
		PagesCount:    336,
		Rating:        4.8,
		CategorySlugs: []string{"drama", "crime"},
	},
	{
		Title:         "1984",
		Description:   "George Orwell's dystopian novel imagines a totalitarian state built on surveillance, propaganda, and fear.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780451524935-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=1984+George+Orwell+book",
		ReleaseDate:   "1949-06-08",
		Author:        "George Orwell",
		PagesCount:    328,
		Rating:        4.7,
		CategorySlugs: []string{"sci-fi", "thriller"},
	},
	{
		Title:         "The Great Gatsby",
		Description:   "A portrait of wealth, longing, and disillusionment in the Jazz Age through the life of Jay Gatsby.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780743273565-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=The+Great+Gatsby+F.+Scott+Fitzgerald+book",
		ReleaseDate:   "1925-04-10",
		Author:        "F. Scott Fitzgerald",
		PagesCount:    180,
		Rating:        4.4,
		CategorySlugs: []string{"drama", "romance"},
	},
	{
		Title:         "Pride and Prejudice",
		Description:   "Jane Austen pairs wit and romance as Elizabeth Bennet and Mr. Darcy learn to see beyond first impressions.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780141439518-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=Pride+and+Prejudice+Jane+Austen+book",
		ReleaseDate:   "1813-01-28",
		Author:        "Jane Austen",
		PagesCount:    432,
		Rating:        4.6,
		CategorySlugs: []string{"romance", "drama"},
	},
	{
		Title:         "The Hobbit",
		Description:   "Bilbo Baggins is swept into a treasure quest that grows into one of fantasy's foundational adventures.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780547928227-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=The+Hobbit+J.R.R.+Tolkien+book",
		ReleaseDate:   "1937-09-21",
		Author:        "J.R.R. Tolkien",
		PagesCount:    300,
		Rating:        4.8,
		CategorySlugs: []string{"fantasy", "adventure"},
	},
	{
		Title:         "The Catcher in the Rye",
		Description:   "Holden Caulfield's restless voice captures alienation, grief, and adolescence in postwar New York.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780316769488-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=The+Catcher+in+the+Rye+J.D.+Salinger+book",
		ReleaseDate:   "1951-07-16",
		Author:        "J.D. Salinger",
		PagesCount:    277,
		Rating:        4.1,
		CategorySlugs: []string{"drama", "psychological"},
	},
	{
		Title:         "The Alchemist",
		Description:   "Paulo Coelho tells a spiritual coming-of-age journey about destiny, faith, and personal legend.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780061122415-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=The+Alchemist+Paulo+Coelho+book",
		ReleaseDate:   "1988-01-01",
		Author:        "Paulo Coelho",
		PagesCount:    208,
		Rating:        4.3,
		CategorySlugs: []string{"adventure", "fantasy"},
	},
	{
		Title:         "Harry Potter and the Sorcerer's Stone",
		Description:   "J.K. Rowling introduces Harry Potter, Hogwarts, and a magical world hidden inside everyday Britain.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780590353427-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=Harry+Potter+and+the+Sorcerer's+Stone+J.K.+Rowling+book",
		ReleaseDate:   "1997-06-26",
		Author:        "J.K. Rowling",
		PagesCount:    309,
		Rating:        4.9,
		CategorySlugs: []string{"fantasy", "family", "adventure"},
	},
	{
		Title:         "The Lord of the Rings",
		Description:   "Tolkien's epic fantasy follows the Fellowship's war against Sauron and the burden of the One Ring.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780544003415-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=The+Lord+of+the+Rings+J.R.R.+Tolkien+book",
		ReleaseDate:   "1954-07-29",
		Author:        "J.R.R. Tolkien",
		PagesCount:    1216,
		Rating:        4.9,
		CategorySlugs: []string{"fantasy", "adventure"},
	},
	{
		Title:         "The Book Thief",
		Description:   "Markus Zusak tells a wartime story narrated by Death about books, family, and resistance in Nazi Germany.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780375842207-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=The+Book+Thief+Markus+Zusak+book",
		ReleaseDate:   "2005-03-14",
		Author:        "Markus Zusak",
		PagesCount:    552,
		Rating:        4.8,
		CategorySlugs: []string{"history", "drama", "war"},
	},
	{
		Title:         "The Da Vinci Code",
		Description:   "A murder at the Louvre launches a fast-moving puzzle thriller through art, religion, and conspiracy.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780307474278-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=The+Da+Vinci+Code+Dan+Brown+book",
		ReleaseDate:   "2003-03-18",
		Author:        "Dan Brown",
		PagesCount:    489,
		Rating:        4.5,
		CategorySlugs: []string{"mystery", "thriller", "suspense"},
	},
	{
		Title:         "Dune",
		Description:   "Frank Herbert blends politics, prophecy, ecology, and war on the desert planet Arrakis.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780441172719-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=Dune+Frank+Herbert+book",
		ReleaseDate:   "1965-08-01",
		Author:        "Frank Herbert",
		PagesCount:    688,
		Rating:        4.8,
		CategorySlugs: []string{"sci-fi", "adventure"},
	},
	{
		Title:         "The Hunger Games",
		Description:   "Katniss Everdeen volunteers for a televised death match in Suzanne Collins' dystopian survival story.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780439023481-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=The+Hunger+Games+Suzanne+Collins+book",
		ReleaseDate:   "2008-09-14",
		Author:        "Suzanne Collins",
		PagesCount:    374,
		Rating:        4.7,
		CategorySlugs: []string{"sci-fi", "adventure", "thriller"},
	},
	{
		Title:         "The Kite Runner",
		Description:   "Khaled Hosseini explores friendship, guilt, and redemption across decades of Afghan history.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9781594631931-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=The+Kite+Runner+Khaled+Hosseini+book",
		ReleaseDate:   "2003-05-29",
		Author:        "Khaled Hosseini",
		PagesCount:    371,
		Rating:        4.7,
		CategorySlugs: []string{"drama", "history"},
	},
	{
		Title:         "Brave New World",
		Description:   "Aldous Huxley imagines a technologically controlled future built on pleasure, conformity, and social engineering.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780060850524-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=Brave+New+World+Aldous+Huxley+book",
		ReleaseDate:   "1932-01-01",
		Author:        "Aldous Huxley",
		PagesCount:    288,
		Rating:        4.3,
		CategorySlugs: []string{"sci-fi", "psychological"},
	},
	{
		Title:         "The Fault in Our Stars",
		Description:   "John Green writes a tender and funny story about love, illness, and growing up too fast.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780142424179-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=The+Fault+in+Our+Stars+John+Green+book",
		ReleaseDate:   "2012-01-10",
		Author:        "John Green",
		PagesCount:    313,
		Rating:        4.5,
		CategorySlugs: []string{"romance", "drama"},
	},
	{
		Title:         "Sapiens",
		Description:   "Yuval Noah Harari surveys the history of humankind from cognitive revolution to modern capitalism.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780062316097-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=Sapiens+Yuval+Noah+Harari+book",
		ReleaseDate:   "2011-01-01",
		Author:        "Yuval Noah Harari",
		PagesCount:    498,
		Rating:        4.7,
		CategorySlugs: []string{"history", "documentary"},
	},
	{
		Title:         "Atomic Habits",
		Description:   "James Clear presents a practical framework for building better routines through small, repeatable changes.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780735211292-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=Atomic+Habits+James+Clear+book",
		ReleaseDate:   "2018-10-16",
		Author:        "James Clear",
		PagesCount:    320,
		Rating:        4.8,
		CategorySlugs: []string{"documentary", "psychological"},
	},
	{
		Title:         "The Silent Patient",
		Description:   "A psychotherapist becomes obsessed with a woman who stopped speaking after allegedly killing her husband.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9781250301697-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=The+Silent+Patient+Alex+Michaelides+book",
		ReleaseDate:   "2019-02-05",
		Author:        "Alex Michaelides",
		PagesCount:    336,
		Rating:        4.4,
		CategorySlugs: []string{"thriller", "mystery", "psychological"},
	},
	{
		Title:         "Educated",
		Description:   "Tara Westover's memoir recounts her path from an isolated upbringing to university life.",
		Type:          "book",
		CoverImage:    "https://covers.openlibrary.org/b/isbn/9780399590504-L.jpg",
		ContentLink:   "https://www.amazon.com/s?k=Educated+Tara+Westover+book",
		ReleaseDate:   "2018-02-20",
		Author:        "Tara Westover",
		PagesCount:    352,
		Rating:        4.7,
		CategorySlugs: []string{"biography", "drama"},
	},
	{
		Title:         "The Dark Knight",
		Description:   "Christopher Nolan pits Batman against the Joker in a crime epic about chaos, sacrifice, and civic order.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/EXeTwQWrcwY/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=EXeTwQWrcwY",
		ReleaseDate:   "2008-07-18",
		Director:      "Christopher Nolan",
		Duration:      152,
		Rating:        9.0,
		CategorySlugs: []string{"action", "crime", "superhero", "thriller"},
	},
	{
		Title:         "Inception",
		Description:   "A thief who steals secrets through dreams takes on a final mission to plant an idea instead.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/YoHD9XEInc0/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=YoHD9XEInc0",
		ReleaseDate:   "2010-07-16",
		Director:      "Christopher Nolan",
		Duration:      148,
		Rating:        8.8,
		CategorySlugs: []string{"sci-fi", "thriller", "action"},
	},
	{
		Title:         "Interstellar",
		Description:   "A team travels through a wormhole to search for humanity's future as Earth becomes less habitable.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/zSWdZVtXT7E/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=zSWdZVtXT7E",
		ReleaseDate:   "2014-11-07",
		Director:      "Christopher Nolan",
		Duration:      169,
		Rating:        8.7,
		CategorySlugs: []string{"sci-fi", "adventure", "drama"},
	},
	{
		Title:         "The Matrix",
		Description:   "Neo discovers reality is a simulation and joins a rebellion against the machines controlling humanity.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/m8e-FF8MsqU/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=m8e-FF8MsqU",
		ReleaseDate:   "1999-03-31",
		Director:      "Lana Wachowski, Lilly Wachowski",
		Duration:      136,
		Rating:        8.7,
		CategorySlugs: []string{"sci-fi", "action"},
	},
	{
		Title:         "Avengers: Endgame",
		Description:   "The Avengers regroup for a final attempt to undo Thanos' devastation and restore half the universe.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/TcMBFSGVi1c/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=TcMBFSGVi1c",
		ReleaseDate:   "2019-04-26",
		Director:      "Anthony Russo, Joe Russo",
		Duration:      181,
		Rating:        8.4,
		CategorySlugs: []string{"action", "superhero", "sci-fi", "adventure"},
	},
	{
		Title:         "Dune",
		Description:   "Denis Villeneuve adapts Frank Herbert's novel into a sweeping struggle over power, prophecy, and Arrakis.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/n9xhJrPXop4/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=n9xhJrPXop4",
		ReleaseDate:   "2021-10-22",
		Director:      "Denis Villeneuve",
		Duration:      155,
		Rating:        8.0,
		CategorySlugs: []string{"sci-fi", "adventure", "drama"},
	},
	{
		Title:         "Joker",
		Description:   "Arthur Fleck's isolation and humiliation spiral into a violent transformation in Gotham City.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/zAGVQLHvwOY/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=zAGVQLHvwOY",
		ReleaseDate:   "2019-10-04",
		Director:      "Todd Phillips",
		Duration:      122,
		Rating:        8.4,
		CategorySlugs: []string{"crime", "drama", "psychological", "thriller"},
	},
	{
		Title:         "Parasite",
		Description:   "Bong Joon-ho's thriller turns class tension into a sharp, unpredictable story of deception and survival.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/SEUXfv87Wpk/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=SEUXfv87Wpk",
		ReleaseDate:   "2019-05-30",
		Director:      "Bong Joon-ho",
		Duration:      132,
		Rating:        8.5,
		CategorySlugs: []string{"thriller", "drama", "comedy"},
	},
	{
		Title:         "Spider-Man: Into the Spider-Verse",
		Description:   "Miles Morales becomes Spider-Man in a visually inventive multiverse adventure.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/g4Hbz2jLxvQ/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=g4Hbz2jLxvQ",
		ReleaseDate:   "2018-12-14",
		Director:      "Bob Persichetti, Peter Ramsey, Rodney Rothman",
		Duration:      117,
		Rating:        8.4,
		CategorySlugs: []string{"animation", "family", "superhero", "adventure"},
	},
	{
		Title:         "Mad Max: Fury Road",
		Description:   "George Miller delivers a relentless desert chase where survival and rebellion collide.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/hEJnMQG9ev8/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=hEJnMQG9ev8",
		ReleaseDate:   "2015-05-15",
		Director:      "George Miller",
		Duration:      120,
		Rating:        8.1,
		CategorySlugs: []string{"action", "adventure", "sci-fi"},
	},
	{
		Title:         "La La Land",
		Description:   "A jazz pianist and aspiring actor chase their dreams while their romance is tested by ambition.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/0pdqf4P9MB8/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=0pdqf4P9MB8",
		ReleaseDate:   "2016-12-09",
		Director:      "Damien Chazelle",
		Duration:      128,
		Rating:        8.0,
		CategorySlugs: []string{"romance", "music", "musical", "drama"},
	},
	{
		Title:         "Whiplash",
		Description:   "A driven drummer and an abusive instructor push each other to the edge in a brutal battle over greatness.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/7d_jQycdQGo/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=7d_jQycdQGo",
		ReleaseDate:   "2014-10-10",
		Director:      "Damien Chazelle",
		Duration:      106,
		Rating:        8.5,
		CategorySlugs: []string{"drama", "music", "psychological"},
	},
	{
		Title:         "The Shawshank Redemption",
		Description:   "Two imprisoned men build a lasting friendship while enduring corruption and hope inside Shawshank prison.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/PLl99DlL6b4/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=PLl99DlL6b4",
		ReleaseDate:   "1994-09-23",
		Director:      "Frank Darabont",
		Duration:      142,
		Rating:        9.3,
		CategorySlugs: []string{"crime", "drama"},
	},
	{
		Title:         "The Godfather",
		Description:   "Francis Ford Coppola chronicles the Corleone family's rise, power, and violence inside organized crime.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/UaVTIH8mujA/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=UaVTIH8mujA",
		ReleaseDate:   "1972-03-24",
		Director:      "Francis Ford Coppola",
		Duration:      175,
		Rating:        9.2,
		CategorySlugs: []string{"crime", "drama"},
	},
	{
		Title:         "Pulp Fiction",
		Description:   "Quentin Tarantino interweaves hitmen, gangsters, and dark comedy into a landmark crime film.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/s7EdQ4FqbhY/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=s7EdQ4FqbhY",
		ReleaseDate:   "1994-10-14",
		Director:      "Quentin Tarantino",
		Duration:      154,
		Rating:        8.9,
		CategorySlugs: []string{"crime", "thriller", "comedy"},
	},
	{
		Title:         "Fight Club",
		Description:   "David Fincher adapts Chuck Palahniuk's novel into a volatile story of identity, masculinity, and revolt.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/qtRKdVHc-cE/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=qtRKdVHc-cE",
		ReleaseDate:   "1999-10-15",
		Director:      "David Fincher",
		Duration:      139,
		Rating:        8.8,
		CategorySlugs: []string{"drama", "thriller", "psychological"},
	},
	{
		Title:         "Spirited Away",
		Description:   "Hayao Miyazaki's animated masterpiece follows Chihiro through a magical bathhouse world.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/ByXuk9QqQkk/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=ByXuk9QqQkk",
		ReleaseDate:   "2001-07-20",
		Director:      "Hayao Miyazaki",
		Duration:      125,
		Rating:        8.6,
		CategorySlugs: []string{"animation", "family", "fantasy", "adventure"},
	},
	{
		Title:         "The Lord of the Rings: The Fellowship of the Ring",
		Description:   "Peter Jackson launches the trilogy with Frodo's first steps toward Mordor and the fall of the Fellowship.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/V75dMMIW2B4/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=V75dMMIW2B4",
		ReleaseDate:   "2001-12-19",
		Director:      "Peter Jackson",
		Duration:      178,
		Rating:        8.8,
		CategorySlugs: []string{"adventure", "fantasy", "family"},
	},
	{
		Title:         "Top Gun: Maverick",
		Description:   "Pete Maverick Mitchell returns to train a new class of pilots for a dangerous mission.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/giXco2jaZ_4/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=giXco2jaZ_4",
		ReleaseDate:   "2022-05-27",
		Director:      "Joseph Kosinski",
		Duration:      130,
		Rating:        8.3,
		CategorySlugs: []string{"action", "drama"},
	},
	{
		Title:         "Oppenheimer",
		Description:   "Christopher Nolan dramatizes J. Robert Oppenheimer's role in the Manhattan Project and its moral fallout.",
		Type:          "movie",
		CoverImage:    "https://img.youtube.com/vi/uYPbbksJxIg/hqdefault.jpg",
		ContentLink:   "https://www.youtube.com/watch?v=uYPbbksJxIg",
		ReleaseDate:   "2023-07-21",
		Director:      "Christopher Nolan",
		Duration:      180,
		Rating:        8.4,
		CategorySlugs: []string{"drama", "history", "thriller", "war"},
	},
	{
		Title:         "Baldur's Gate 3",
		Description:   "Larian's RPG adapts Dungeons & Dragons into a choice-driven adventure across the Forgotten Realms.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1086940/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/1086940/Baldurs_Gate_3/",
		ReleaseDate:   "2023-08-03",
		Developer:     "Larian Studios",
		Platform:      "Windows, macOS",
		Rating:        9.6,
		CategorySlugs: []string{"adventure", "fantasy"},
	},
	{
		Title:         "Elden Ring",
		Description:   "FromSoftware and George R.R. Martin deliver a vast open-world action RPG set in the Lands Between.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1245620/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/1245620/ELDEN_RING/",
		ReleaseDate:   "2022-02-25",
		Developer:     "FromSoftware, Inc.",
		Platform:      "Windows",
		Rating:        9.5,
		CategorySlugs: []string{"action", "adventure", "fantasy"},
	},
	{
		Title:         "Hades",
		Description:   "Supergiant's roguelike sends Zagreus through the Underworld in fast, stylish escape attempts.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1145360/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/1145360/Hades/",
		ReleaseDate:   "2020-09-17",
		Developer:     "Supergiant Games",
		Platform:      "Windows, macOS",
		Rating:        9.3,
		CategorySlugs: []string{"action", "fantasy"},
	},
	{
		Title:         "Portal 2",
		Description:   "Valve expands its puzzle classic with portal mechanics, sharp writing, and an all-time great co-op mode.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/620/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/620/Portal_2/",
		ReleaseDate:   "2011-04-19",
		Developer:     "Valve",
		Platform:      "Windows, macOS, Linux",
		Rating:        9.6,
		CategorySlugs: []string{"adventure", "comedy", "sci-fi"},
	},
	{
		Title:         "The Witcher 3: Wild Hunt",
		Description:   "Geralt's hunt for Ciri unfolds across a massive fantasy world filled with monsters, war, and political intrigue.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/292030/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/292030/The_Witcher_3_Wild_Hunt/",
		ReleaseDate:   "2015-05-19",
		Developer:     "CD PROJEKT RED",
		Platform:      "Windows",
		Rating:        9.7,
		CategorySlugs: []string{"adventure", "fantasy", "drama"},
	},
	{
		Title:         "Stardew Valley",
		Description:   "ConcernedApe turns farming, friendship, and small-town routines into a deeply replayable life sim.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/413150/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/413150/Stardew_Valley/",
		ReleaseDate:   "2016-02-26",
		Developer:     "ConcernedApe",
		Platform:      "Windows, macOS, Linux",
		Rating:        9.4,
		CategorySlugs: []string{"family", "romance", "adventure"},
	},
	{
		Title:         "Hollow Knight",
		Description:   "Team Cherry's atmospheric metroidvania mixes hard combat, exploration, and haunting worldbuilding.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/367520/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/367520/Hollow_Knight/",
		ReleaseDate:   "2017-02-24",
		Developer:     "Team Cherry",
		Platform:      "Windows, macOS, Linux",
		Rating:        9.4,
		CategorySlugs: []string{"action", "adventure", "fantasy"},
	},
	{
		Title:         "Red Dead Redemption 2",
		Description:   "Rockstar's western epic follows Arthur Morgan through the decline of the Van der Linde gang.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1174180/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/1174180/Red_Dead_Redemption_2/",
		ReleaseDate:   "2019-12-05",
		Developer:     "Rockstar Games",
		Platform:      "Windows",
		Rating:        9.5,
		CategorySlugs: []string{"action", "adventure", "western", "drama"},
	},
	{
		Title:         "Cyberpunk 2077",
		Description:   "Night City becomes the backdrop for a story about identity, power, and high-tech survival.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1091500/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/1091500/Cyberpunk_2077/",
		ReleaseDate:   "2020-12-10",
		Developer:     "CD PROJEKT RED",
		Platform:      "Windows",
		Rating:        8.6,
		CategorySlugs: []string{"sci-fi", "action", "thriller"},
	},
	{
		Title:         "Celeste",
		Description:   "A demanding platformer about climbing a mountain and confronting anxiety with patience and persistence.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/504230/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/504230/Celeste/",
		ReleaseDate:   "2018-01-25",
		Developer:     "Extremely OK Games, Ltd.",
		Platform:      "Windows, macOS, Linux",
		Rating:        9.2,
		CategorySlugs: []string{"adventure", "drama"},
	},
	{
		Title:         "Slay the Spire",
		Description:   "Mega Crit blends deckbuilding and roguelike runs into one of PC gaming's most replayable strategy games.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/646570/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/646570/Slay_the_Spire/",
		ReleaseDate:   "2019-01-23",
		Developer:     "Mega Crit",
		Platform:      "Windows, macOS, Linux",
		Rating:        9.2,
		CategorySlugs: []string{"adventure", "fantasy"},
	},
	{
		Title:         "Disco Elysium - The Final Cut",
		Description:   "A detective RPG where conversation, ideology, and internal conflict matter more than combat.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/632470/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/632470/Disco_Elysium__The_Final_Cut/",
		ReleaseDate:   "2019-10-15",
		Developer:     "ZA/UM",
		Platform:      "Windows, macOS",
		Rating:        9.2,
		CategorySlugs: []string{"mystery", "drama", "psychological"},
	},
	{
		Title:         "Sekiro: Shadows Die Twice",
		Description:   "FromSoftware shifts its precision combat into a faster shinobi adventure set in a mythic Sengoku Japan.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/814380/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/814380/Sekiro_Shadows_Die_Twice/",
		ReleaseDate:   "2019-03-22",
		Developer:     "FromSoftware, Inc.",
		Platform:      "Windows",
		Rating:        9.4,
		CategorySlugs: []string{"action", "adventure", "fantasy"},
	},
	{
		Title:         "Persona 5 Royal",
		Description:   "Stylish turn-based battles and social simulation drive Atlus' story of rebellion inside modern Tokyo.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1687950/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/1687950/Persona_5_Royal/",
		ReleaseDate:   "2022-10-21",
		Developer:     "ATLUS",
		Platform:      "Windows",
		Rating:        9.3,
		CategorySlugs: []string{"adventure", "fantasy", "drama"},
	},
	{
		Title:         "God of War",
		Description:   "Kratos and Atreus travel through Norse realms in a character-driven action adventure.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1593500/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/1593500/God_of_War/",
		ReleaseDate:   "2022-01-14",
		Developer:     "Santa Monica Studio",
		Platform:      "Windows",
		Rating:        9.4,
		CategorySlugs: []string{"action", "adventure", "fantasy"},
	},
	{
		Title:         "Resident Evil 4",
		Description:   "Capcom reimagines the survival horror classic with tighter action and modern production values.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/2050650/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/2050650/Resident_Evil_4/",
		ReleaseDate:   "2023-03-24",
		Developer:     "CAPCOM Co., Ltd.",
		Platform:      "Windows",
		Rating:        9.1,
		CategorySlugs: []string{"action", "horror", "suspense"},
	},
	{
		Title:         "Monster Hunter: World",
		Description:   "Track giant monsters, forge better gear, and loop back into even harder hunts in Capcom's breakout hit.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/582010/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/582010/Monster_Hunter_World/",
		ReleaseDate:   "2018-08-09",
		Developer:     "CAPCOM Co., Ltd.",
		Platform:      "Windows",
		Rating:        9.0,
		CategorySlugs: []string{"action", "adventure", "fantasy"},
	},
	{
		Title:         "Terraria",
		Description:   "Re-Logic's sandbox classic turns mining, crafting, combat, and exploration into an endless adventure loop.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/105600/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/105600/Terraria/",
		ReleaseDate:   "2011-05-16",
		Developer:     "Re-Logic",
		Platform:      "Windows, macOS, Linux",
		Rating:        9.4,
		CategorySlugs: []string{"adventure", "fantasy", "family"},
	},
	{
		Title:         "Euro Truck Simulator 2",
		Description:   "SCS Software turns long-haul trucking into a calm, satisfying road-trip management game.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/227300/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/227300/Euro_Truck_Simulator_2/",
		ReleaseDate:   "2012-10-18",
		Developer:     "SCS Software",
		Platform:      "Windows, macOS, Linux",
		Rating:        8.8,
		CategorySlugs: []string{"adventure", "documentary"},
	},
	{
		Title:         "Vampire Survivors",
		Description:   "Poncle distills survival action into short, chaotic runs packed with unlocks and screen-filling builds.",
		Type:          "game",
		CoverImage:    "https://shared.cloudflare.steamstatic.com/store_item_assets/steam/apps/1794680/header.jpg",
		ContentLink:   "https://store.steampowered.com/app/1794680/Vampire_Survivors/",
		ReleaseDate:   "2022-10-20",
		Developer:     "poncle",
		Platform:      "Windows, macOS, Linux",
		Rating:        8.9,
		CategorySlugs: []string{"action", "fantasy", "horror"},
	},
}

func SeedItems() {
	if DbConn == nil {
		panic("database is not connected")
	}

	tx := DbConn.Begin()
	if tx.Error != nil {
		panic(fmt.Sprintf("item seed transaction failed: %v", tx.Error))
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	categoryMap, err := loadSeedCategories(tx)
	if err != nil {
		tx.Rollback()
		panic(fmt.Sprintf("item seed category load failed: %v", err))
	}

	createdCount := 0
	updatedCount := 0

	for _, entry := range defaultItems {
		categories, err := categoriesForSeed(entry.CategorySlugs, categoryMap)
		if err != nil {
			tx.Rollback()
			panic(fmt.Sprintf("item seed categories failed for %s: %v", entry.Title, err))
		}

		record := entry.toModel()

		var existing models.Item
		result := tx.Where("title = ? AND type = ?", entry.Title, entry.Type).First(&existing)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			tx.Rollback()
			panic(fmt.Sprintf("item seed query failed for %s: %v", entry.Title, result.Error))
		}

		if result.RowsAffected == 0 {
			if err := tx.Create(&record).Error; err != nil {
				tx.Rollback()
				panic(fmt.Sprintf("item seed create failed for %s: %v", entry.Title, err))
			}
			createdCount++
		} else {
			existing.Title = record.Title
			existing.Description = record.Description
			existing.Type = record.Type
			existing.CoverImage = record.CoverImage
			existing.ContentLink = record.ContentLink
			existing.ReleaseDate = record.ReleaseDate
			existing.Author = record.Author
			existing.Director = record.Director
			existing.Developer = record.Developer
			existing.Duration = record.Duration
			existing.PagesCount = record.PagesCount
			existing.Platform = record.Platform
			existing.Rating = record.Rating

			if err := tx.Save(&existing).Error; err != nil {
				tx.Rollback()
				panic(fmt.Sprintf("item seed update failed for %s: %v", entry.Title, err))
			}

			record = existing
			updatedCount++
		}

		if err := tx.Model(&record).Association("Categories").Replace(categories); err != nil {
			tx.Rollback()
			panic(fmt.Sprintf("item seed category replace failed for %s: %v", entry.Title, err))
		}
	}

	if err := tx.Commit().Error; err != nil {
		panic(fmt.Sprintf("item seed commit failed: %v", err))
	}

	log.Printf("item seed: ensured %d items (%d created, %d updated)", len(defaultItems), createdCount, updatedCount)
}

func loadSeedCategories(tx *gorm.DB) (map[string]models.Category, error) {
	var categories []models.Category
	if err := tx.Find(&categories).Error; err != nil {
		return nil, err
	}

	categoryMap := make(map[string]models.Category, len(categories))
	for _, category := range categories {
		categoryMap[category.Slug] = category
	}

	return categoryMap, nil
}

func categoriesForSeed(slugs []string, categoryMap map[string]models.Category) ([]models.Category, error) {
	categories := make([]models.Category, 0, len(slugs))
	for _, slug := range slugs {
		category, ok := categoryMap[slug]
		if !ok {
			return nil, fmt.Errorf("missing category slug %q", slug)
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func (s seedItem) toModel() models.Item {
	record := models.Item{
		Title:       s.Title,
		Description: s.Description,
		Type:        s.Type,
		CoverImage:  s.CoverImage,
		ReleaseDate: mustParseSeedDate(s.ReleaseDate),
		Rating:      s.Rating,
	}

	if s.ContentLink != "" {
		record.ContentLink = stringPtr(s.ContentLink)
	}
	if s.Author != "" {
		record.Author = stringPtr(s.Author)
	}
	if s.Director != "" {
		record.Director = stringPtr(s.Director)
	}
	if s.Developer != "" {
		record.Developer = stringPtr(s.Developer)
	}
	if s.Platform != "" {
		record.Platform = stringPtr(s.Platform)
	}
	if s.Duration > 0 {
		record.Duration = uintPtr(s.Duration)
	}
	if s.PagesCount > 0 {
		record.PagesCount = uintPtr(s.PagesCount)
	}

	return record
}

func mustParseSeedDate(value string) time.Time {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		panic(fmt.Sprintf("invalid seed date %q: %v", value, err))
	}

	return parsed
}

func stringPtr(value string) *string {
	return &value
}

func uintPtr(value uint) *uint {
	return &value
}
