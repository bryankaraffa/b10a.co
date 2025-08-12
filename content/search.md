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
    <p>Loading posts...</p>
  </div>
</div>

<script>
// Client-side search functionality with dynamic post loading
const searchInput = document.getElementById('search-input');
const searchResults = document.getElementById('search-results');
let posts = [];

// Load posts from the generated JSON index
async function loadPosts() {
  try {
    const response = await fetch('/index.json');
    posts = await response.json();
    searchResults.innerHTML = '<p>Start typing to search through blog posts...</p>';
  } catch (error) {
    console.error('Error loading posts:', error);
    searchResults.innerHTML = '<p>Error loading posts. Please refresh the page.</p>';
  }
}

function performSearch(query) {
  if (query.length < 2) {
    searchResults.innerHTML = '<p>Start typing to search through blog posts...</p>';
    return;
  }

  const searchTerm = query.toLowerCase();
  const results = posts.filter(post => 
    post.title.toLowerCase().includes(searchTerm) ||
    post.excerpt.toLowerCase().includes(searchTerm) ||
    (post.tags && post.tags.some(tag => tag.toLowerCase().includes(searchTerm)))
  );

  if (results.length === 0) {
    searchResults.innerHTML = '<p>No posts found matching your search.</p>';
    return;
  }

  let html = `<p>Found ${results.length} post${results.length === 1 ? '' : 's'}:</p>`;
  html += '<div>';
  
  results.forEach(post => {
    const tags = post.tags || [];
    html += `
      <div style="margin-bottom: 20px; padding: 15px; border: 1px solid #e0e0e0; border-radius: 5px;">
        <h3 style="margin: 0 0 10px 0;"><a href="${post.url}" style="color: #2563eb; text-decoration: none;">${post.title}</a></h3>
        <p style="margin: 0 0 10px 0; color: #666;">${post.excerpt}</p>
        ${tags.length > 0 ? `
        <div style="font-size: 0.9em; color: #888;">
          Tags: ${tags.map(tag => `<span style="background: #f0f0f0; padding: 2px 6px; margin-right: 5px; border-radius: 3px;">${tag}</span>`).join('')}
        </div>
        ` : ''}
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

// Load posts when page loads
loadPosts();
</script>