// src/app/components/dashboard/dashboard.component.ts
import { Component, OnInit } from '@angular/core';
import {SidebarComponent} from '../sidebar/sidebar.component';

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  imports: [
    SidebarComponent
  ],
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit {
  user = {
    name: 'Max Mustermann',
    avatar: 'assets/user-avatar.png',
    level: 5,
    xp: 2750,
    nextLevelXp: 3000
  };

  recentCourse = {
    title: 'JavaScript Grundlagen',
    progress: 65,
    lastLesson: 'Funktionen und Callbacks',
    imageUrl: 'assets/js-course.jpg'
  };

  recentChat = {
    title: 'Hilfe bei Arrays',
    lastMessage: 'Kannst du mir erklären, wie map() funktioniert?',
    time: '14:25',
    unread: 2
  };

  recommendedCourses = [
    {
      id: 1,
      title: 'React für Fortgeschrittene',
      duration: '8 Stunden',
      level: 'Fortgeschritten',
      imageUrl: 'assets/react-course.jpg'
    },
    {
      id: 2,
      title: 'CSS Flexbox & Grid',
      duration: '5 Stunden',
      level: 'Mittel',
      imageUrl: 'assets/css-course.jpg'
    },
    {
      id: 3,
      title: 'TypeScript Basics',
      duration: '6 Stunden',
      level: 'Anfänger',
      imageUrl: 'assets/typescript-course.jpg'
    }
  ];

  learningPath = {
    title: 'Full-Stack Entwickler',
    progress: 42,
    nextMilestone: 'MongoDB Einführung',
    completedModules: 8,
    totalModules: 19
  };

  currentStreak = 7; // Tage
  totalCourses = 12;
  completedCourses = 5;
  totalExercises = 87;
  completedExercises = 53;

  constructor() { }

  ngOnInit(): void {
    // Hier könnten API-Aufrufe für Benutzerdaten, Kurse usw. erfolgen
  }
}
