import { Component } from '@angular/core';

@Component({
  selector: 'app-ai-chat',
  imports: [],
  templateUrl: './ai-chat.component.html',
  styleUrl: './ai-chat.component.scss'
})
export class AiChatComponent {
  //.js
  const searchInput = document.querySelector('.search-input');
  searchInput.addEventListener('input', (e) => {
  const searchTerm = e.target.value.toLowerCase();
  const cards = document.querySelectorAll('.course-card');

  cards.forEach(card => {
  const title = card.querySelector('.course-title').textContent.toLowerCase();
  const description = card.querySelector('.course-description').textContent.toLowerCase();
  const tags = Array.from(card.querySelectorAll('.tag')).map(tag => tag.textContent.toLowerCase());

  if (title.includes(searchTerm) ||
  description.includes(searchTerm) ||
  tags.some(tag => tag.includes(searchTerm))) {
  card.style.display = 'block';
} else {
  card.style.display = 'none';
}
});
});

// Filter functionality
const filterButtons = document.querySelectorAll('.filter-button');
filterButtons.forEach(button => {
  button.addEventListener('click', () => {
    filterButtons.forEach(btn => btn.classList.remove('active'));
    button.classList.add('active');
  });
});

// Sidebar item active state
const sidebarItems = document.querySelectorAll('.sidebar-item');
sidebarItems.forEach(item => {
  item.addEventListener('click', () => {
    sidebarItems.forEach(i => i.classList.remove('active'));
    item.classList.add('active');
  });
});
}
