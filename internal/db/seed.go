package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	"github.com/ozarabal/goSocial/internal/store"
)

var usernames = []string{"user123", "alphaWolf","techGuru","fastCoder",
  "skyDreamer","mountainKing","bluePhoenix","digitalNomad","shadowHunter",
  "fireBlade","cosmicVoyager","goldenEagle","nightHawk","silentStorm","ironFist",
  "quantumWizard","starGazer","pixelArtist","ninjaCoder","darkKnight",
  "crimsonArrow","electricPulse","codeBreaker","lunarEclipse","oceanExplorer",
  "wildFox","phantomCoder","solarFlare","stormRider","cyberGlitch",
}

var titles = []string{
  "The Future of Artificial Intelligence",
  "10 Tips for Staying Productive While Working from Home",
  "Top 5 Destinations for Adventure Seekers",
  "The Benefits of Meditation for Mental Health",
  "How to Build a Personal Brand Online",
  "The Best Programming Languages to Learn in 2025",
  "Understanding Cryptocurrency: A Beginner's Guide",
  "Healthy Eating on a Budget",
  "How to Create a Successful YouTube Channel",
  "The Rise of Electric Vehicles",
  "Mastering Time Management for Students",
  "Why You Should Start Journaling",
  "The Science Behind Good Sleep",
  "Top 7 Gadgets for Tech Enthusiasts in 2025",
  "How to Start a Side Hustle",
  "The Basics of Sustainable Living",
  "Effective Strategies for Learning a New Language",
  "The Evolution of Social Media",
  "How to Improve Your Public Speaking Skills",
  "5 Exercises for a Stronger Core",
};

var contents = []string{
  "AI is transforming the way we live, work, and interact. Here's a look at the latest advancements and their potential impact on society.",
  "Remote work can be challenging. These tips will help you stay focused and maintain a healthy work-life balance.",
  "Explore these breathtaking locations if you crave adventure and outdoor thrills.",
  "Meditation has been shown to reduce stress and improve focus. Discover how you can start your meditation journey today.",
  "Your online presence can open new opportunities. Learn the steps to create a personal brand that stands out.",
  "Discover which programming languages are in demand and why you should consider learning them.",
  "Cryptocurrencies are reshaping the financial landscape. Here's a beginner-friendly overview to get started.",
  "Eating healthy doesn't have to break the bank. Check out these tips for nutritious meals at an affordable price.",
  "Building a YouTube channel takes strategy and dedication. Here’s how you can grow your audience and create engaging content.",
  "Electric vehicles are becoming more popular than ever. Learn about the benefits and challenges of this growing trend.",
  "Balancing school, hobbies, and personal life can be tough. These time management strategies will help students thrive.",
  "Journaling can improve mental clarity and track personal growth. Learn how to start and stick with it.",
  "Quality sleep is essential for health and productivity. Understand the science and tips to improve your sleep habits.",
  "Explore the latest gadgets that are redefining technology and making life easier for tech lovers.",
  "Turning your passion into profit is easier than you think. Follow these steps to start your side hustle today.",
  "Living sustainably helps the environment and saves money. Here’s how you can start making eco-friendly choices.",
  "Learning a new language can be fun and rewarding. Use these proven strategies to speed up the process.",
  "From MySpace to TikTok, social media has evolved rapidly. Take a look at how it has shaped our communication.",
  "Public speaking can be daunting, but with practice and the right techniques, you can become a confident speaker.",
  "A strong core improves posture and stability. Incorporate these simple exercises into your fitness routine.",
  };

  var comments = []string{
    "This is a great post! Thanks for sharing.",
  "I completely agree with your point of view.",
  "This topic is so relevant right now.",
  "Thanks for the detailed explanation. It really helped!",
  "Interesting perspective. I never thought of it that way.",
  "Can you share more details on this?",
  "Amazing content as always. Keep it up!",
  "I learned so much from this post. Thank you!",
  "This is exactly what I was looking for!",
  "Do you have any resources to dive deeper into this topic?",
  "I appreciate how clearly you explained everything.",
  "Well written and very informative. Thanks!",
  "I have a question about one of the points you mentioned.",
  "Great tips! I'll definitely try this out.",
  "This is such a useful post. Thank you for sharing!",
  "I'm bookmarking this for future reference.",
  "I love how you broke this down into simple steps.",
  "This really resonated with me. Thanks for sharing!",
  "Can you do a follow-up post on this topic?",
  "Your posts are always so insightful. Keep it coming!",
  "This was very eye-opening. Thanks for sharing your knowledge.",
  "I hadn't considered this before. Thanks for the fresh perspective!",
  "Great content! Looking forward to more posts like this.",
  "This is a game-changer for me. Thanks for writing it!",
  "You explained this so well. Even a beginner can understand it!",
  "I shared this with my friends. It's too good not to share!",
  "I tried this, and it really works! Thanks!",
  "Do you have any examples to illustrate this further?",
  "This is such an underrated topic. Glad you wrote about it!",
  "Your writing style is so engaging. I couldn't stop reading!",
  }

  var tags = []string{
    "Self Improvement", "Minimalism", "Health", "Travel", "Mindfulness",
	"Productivity", "Home Office", "Digital Detox", "Gardening", "DIY",
	"Yoga", "Sustainability", "Time Management", "Nature", "Cooking",
	"Fitness", "Personal Finance", "Writing", "Mental Health", "Learning",
  }


func Seed(store store.Storage, db *sql.DB){
	ctx := context.Background()

	users := generateUsers(30)
  tx, _ :=  db.BeginTx(ctx, nil)

	for _, user := range users{
		if err := store.Users.Create(ctx,tx, user); err != nil{
      _ = tx.Rollback()
			log.Println("Error creating user:", err)
			return	
		}
	}
  tx.Commit()
	
	posts := generatePosts(200, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Println("Error creating post:", err)
			return
		}
	}

  comments := generateComments(500, users, posts)
  for _,comment := range comments{
    if err := store.Comments.Create(ctx, comment); err != nil {
        log.Println("Error creating comment:", err)
        return
    }
  }

  log.Println("Seeding complete")
}

func generateUsers(num int) []*store.User{
	users := make([]*store.User, num)

	for i:=0; i<num; i++ {
		users[i] = &store.User{
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email: usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "example.com",
		}
	}
	return users
}

func generatePosts(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)
	for i := 0; i < num; i++ {
		user := users[rand.Intn(len(users))]

		posts[i] = &store.Post{
			UserId:  user.ID,
			Title:   titles[rand.Intn(len(titles))],
			Content: titles[rand.Intn(len(contents))],
			Tags: []string{
				tags[rand.Intn(len(tags))],
				tags[rand.Intn(len(tags))],
			},
		}
	}

	return posts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	cms := make([]*store.Comment, num)
	for i := 0; i < num; i++ {
		cms[i] = &store.Comment{
			PostId:  posts[rand.Intn(len(posts))].ID,
			UserId:  users[rand.Intn(len(users))].ID,
			Content: comments[rand.Intn(len(comments))],
		}
	}
	return cms
}
