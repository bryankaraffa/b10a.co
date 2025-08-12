---
title: "Search"
date: 2023-01-01
draft: false
type: page
nodate: true
hidemeta: true
---

# Search Posts

<div id="search-container">
  <input type="text" id="search-input" placeholder="Search posts..." style="width: 100%; padding: 10px; margin-bottom: 20px; border: 1px solid #ccc; border-radius: 4px; font-size: 16px;">
  
  <div id="search-results">
    <p>Start typing to search through blog posts...</p>
  </div>
</div>

<script>
// Simple client-side search functionality
const searchInput = document.getElementById('search-input');
const searchResults = document.getElementById('search-results');

// This would be better loaded from a JSON index, but for simplicity we'll hardcode some posts
const posts = [
  {
    title: "266 Spam Posts Later: Migrating the Guestbook from Staticman to a Custom Go Solution",
    url: "/posts/2025/08/266-spam-posts-later-migrating-guestbook-from-staticman-to-custom-go-solution/",
    excerpt: "My Guestbook page, which contains a simple form accepting posts without any authentication required...",
    tags: ["hugo", "golang", "guestbook", "gin", "spam", "staticman", "open-source"]
  },
  {
    title: "The Bots Are Winning, But Have Not Won", 
    url: "/posts/2025/03/the-bots-are-winning-but-have-not-won/",
    excerpt: "In my previous post, I shared my experiment of adding a public guestbook to my Hugo site using Staticman...",
    tags: ["hugo", "open source", "staticman", "spam"]
  },
  {
    title: "Adding a Guestbook to Hugo website using Staticman",
    url: "/posts/2023/06/guestbook-for-hugo-using-staticman/",
    excerpt: "This website is powered by Hugo which is a static site generator, and static websites typically do not support user-generated content...",
    tags: ["hugo", "staticman", "guestbook"]
  },
  {
    title: "Hello World",
    url: "/posts/2023/01/hello-world/",
    excerpt: "b10a.co is online. Website build with Hugo and hosted on GitHub Pages. Costs $0 per month...",
    tags: ["free"]
  },
  {
    title: "Using RTL-SDR USB devices in WSL on Windows",
    url: "/posts/2023/12/rtlsdr-in-wsl/",
    excerpt: "There is a way to connect USB to WSL2. Here's a loosely noted outline for how I was able to get my RTL-SDR USB device working in a Docker...",
    tags: ["rtl-sdr", "wsl", "windows", "linux"]
  },
  {
    title: "A Cost Analysis of GeForce NOW and Game Streaming versus \"Bare Metal\" Laptop or PC for Gaming",
    url: "/posts/2023/11/geforce-now-cost-analysis/",
    excerpt: "I am gamer. That's how I got into IT and coding as a kid. I sometimes use gaming servers or use-cases to explore the limits...",
    tags: ["gaming", "cloud", "opinion"]
  }
];

function performSearch(query) {
  if (query.length < 2) {
    searchResults.innerHTML = '<p>Start typing to search through blog posts...</p>';
    return;
  }

  const searchTerm = query.toLowerCase();
  const results = posts.filter(post => 
    post.title.toLowerCase().includes(searchTerm) ||
    post.excerpt.toLowerCase().includes(searchTerm) ||
    post.tags.some(tag => tag.toLowerCase().includes(searchTerm))
  );

  if (results.length === 0) {
    searchResults.innerHTML = '<p>No posts found matching your search.</p>';
    return;
  }

  let html = `<p>Found ${results.length} post${results.length === 1 ? '' : 's'}:</p>`;
  html += '<div>';
  
  results.forEach(post => {
    html += `
      <div style="margin-bottom: 20px; padding: 15px; border: 1px solid #e0e0e0; border-radius: 5px;">
        <h3 style="margin: 0 0 10px 0;"><a href="${post.url}" style="color: #2563eb; text-decoration: none;">${post.title}</a></h3>
        <p style="margin: 0 0 10px 0; color: #666;">${post.excerpt}</p>
        <div style="font-size: 0.9em; color: #888;">
          Tags: ${post.tags.map(tag => `<span style="background: #f0f0f0; padding: 2px 6px; margin-right: 5px; border-radius: 3px;">${tag}</span>`).join('')}
        </div>
      </div>
    `;
  });
  
  html += '</div>';
  searchResults.innerHTML = html;
}

// Add search functionality
searchInput.addEventListener('input', (e) => {
  performSearch(e.target.value);
});

// Add search on enter
searchInput.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') {
    e.preventDefault();
    performSearch(e.target.value);
  }
});
</script>