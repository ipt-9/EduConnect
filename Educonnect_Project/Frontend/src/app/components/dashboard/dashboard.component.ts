// src/app/components/dashboard/dashboard.component.ts
import { Component, OnInit } from '@angular/core';
import { SidebarComponent } from '../sidebar/sidebar.component';
import { CommonModule } from '@angular/common';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Router } from '@angular/router';

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  imports: [
    SidebarComponent,
    CommonModule
  ],
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit {
  sidebarExpanded = true;

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

  learningPath = {
    title: 'Full-Stack Entwickler',
    progress: 42,
    nextMilestone: 'MongoDB Einführung',
    completedModules: 8,
    totalModules: 19
  };

  currentStreak = 7;
  totalCourses = 12;
  completedCourses = 5;
  totalExercises = 87;
  completedExercises = 53;

  myCourses: any[] = []; // 🆕 aus API geladen

  constructor(private http: HttpClient, private router: Router) {}

  ngOnInit(): void {
    this.loadMyCourses();
  }

  onSidebarExpand(value: boolean): void {
    this.sidebarExpanded = value;
  }

  loadMyCourses(): void {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders().set('Authorization', `Bearer ${token}`);

    this.http.get<any[]>('http://localhost:8080/my-courses', { headers }).subscribe({
      next: (data) => {
        this.myCourses = data;
        console.log('📚 Meine Kurse geladen:', data);
      },
      error: (err) => {
        console.error('❌ Fehler beim Laden der Kurse:', err);
      }
    });
  }

  openCourse(courseId: number): void {
    localStorage.setItem('activeCourseId', courseId.toString());
    this.router.navigate(['/taskslist']);
  }

  getCourseImage(language: string): string {
    switch (language.toLowerCase()) {
      case 'python':
        return 'assets/img/python-cover.png';
      case 'javascript':
        return 'assets/img/javascript-cover.png';
      case 'typescript':
        return 'assets/img/typescript-cover.png';
      case 'java':
        return 'assets/img/java-cover.png';
      default:
        return 'assets/img/default-course-cover.png';
    }
  }
}
