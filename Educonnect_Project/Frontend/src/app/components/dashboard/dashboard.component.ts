// src/app/components/dashboard/dashboard.component.ts
import { Component, OnInit } from '@angular/core';
import { SidebarComponent } from '../sidebar/sidebar.component';
import { CommonModule } from '@angular/common';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Router } from '@angular/router';
import { RouterModule } from '@angular/router';


@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  imports: [
    SidebarComponent,
    CommonModule,
    RouterModule
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
  public recentCourse = {
    title: '',
    progress: 0,
    lastLesson: '',
    taskDescription: '',
    imageUrl: '',
    courseId: false
  };
  public dashboardOverview = {
    lastMessageText: '',
    lastMessageCreatedAt: '',
    lastMessageGroupId: 0,
    nextPendingTaskTitle: '',
    nextPendingTaskId: 0
  };
  hasSubscription: boolean = false;

  checkSubscriptionStatus(): void {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders().set('Authorization', `Bearer ${token}`);

    this.http.get<{ has_subscription: boolean }>('https://api.educonnect-bmsd22a.bbzwinf.ch/subscription-status', { headers }).subscribe({
      next: (data) => {
        console.log('🔒 Abo-Status geladen:', data.has_subscription);
        this.hasSubscription = data.has_subscription;
      },
      error: (err) => {
        console.error('❌ Fehler beim Prüfen des Abo-Status:', err);
      }
    });
  }


  loadDashboardOverview(): void {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders().set('Authorization', `Bearer ${token}`);

    this.http.get<any>('https://api.educonnect-bmsd22a.bbzwinf.ch/dashboard-overview', { headers }).subscribe({
      next: (data) => {
        console.log('📊 Dashboard Overview:', data);
        this.dashboardOverview.lastMessageText = data.last_message_text || 'Keine neue Nachricht';
        this.dashboardOverview.lastMessageCreatedAt = this.formatDateToZurich(data.last_message_created_at);
        this.dashboardOverview.lastMessageGroupId = data.last_message_group_id || 0;
        this.dashboardOverview.nextPendingTaskTitle = data.next_pending_task_title;
        this.dashboardOverview.nextPendingTaskId = data.next_pending_task_id;
      },
      error: (err) => {
        console.error('❌ Fehler beim Laden der Dashboard-Übersicht:', err);
      }
    });
  }

  loadLastVisitedCourseAndTask(): void {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders().set('Authorization', `Bearer ${token}`);

    this.http.get<any>('https://api.educonnect-bmsd22a.bbzwinf.ch/last-course', { headers }).subscribe({
      next: (data) => {
        console.log('🕹️ Letzter besuchter Kurs und Aufgabe:', data);
        this.recentCourse.title = data.course_title;
        this.recentCourse.progress = data.progress_percent;
        this.recentCourse.lastLesson = data.task_title;
        this.recentCourse.taskDescription = data.task_description;
        this.recentCourse.imageUrl = this.getCourseImage(data.language);
        this.recentCourse.courseId = data.course_id;
        },
      error: (err) => {
        console.error('❌ Fehler beim Laden des letzten Kurses:', err);
      }
    });
  }

  openTask(taskId: number): void {
    if (taskId) {
      localStorage.setItem('activeTaskId', taskId.toString());
      this.router.navigate(['/codingSpace']);
    } else {
      console.error('⚠️ Keine offene Aufgabe gefunden.');
    }
  }


  recentChat = {
    title: 'Letzter Chat',
    lastMessage: '',
    time: ''
  };


  learningPath = {
    title: 'Full-Stack Entwickler',
    progress: 42,
    nextMilestone: 'MongoDB Einführung',
    completedModules: 8,
    totalModules: 19
  };

  statistics = {
    completedCourses: 0,
    totalCourses: 3,
    completedExercises: 0,
    totalExercises: 30,
    currentStreak: 'Coming Soon'
  };


  myCourses: any[] = []; // 🆕 aus API geladen

  constructor(private http: HttpClient, private router: Router) {}

  ngOnInit(): void {
    this.checkSubscriptionStatus();
    this.loadMyCourses();
    this.loadUser();
    this.loadLastVisitedCourseAndTask();
    this.loadDashboardOverview();
    this.loadUserStatistics();
  }

  onSidebarExpand(value: boolean): void {
    this.sidebarExpanded = value;
  }
  navigateToPayment(): void {
    this.router.navigate(['/payment']);
  }
  loadMyCourses(): void {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders().set('Authorization', `Bearer ${token}`);

    this.http.get<any[]>('https://api.educonnect-bmsd22a.bbzwinf.ch/my-courses', { headers }).subscribe({
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
  loadUser(): void { // 🆕 Neue Funktion für den Benutzernamen
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders().set('Authorization', `Bearer ${token}`);

    this.http.get<{ username: string }>('https://api.educonnect-bmsd22a.bbzwinf.ch/me', {headers}).subscribe({
      next: (data) => {
        this.user.name = data.username;
        console.log('🙋 Benutzer geladen:', data.username);
      },
      error: (err) => {
        console.error('❌ Fehler beim Laden des Benutzers:', err);
      }
    });
  }
  continueCourse(): void {
    if (this.recentCourse && this.recentCourse.courseId) {
      localStorage.setItem('activeCourseId', this.recentCourse.courseId.toString());
      this.router.navigate(['/taskslist']);
    } else {
      console.error('⚠️ Kein Kurs zum Fortsetzen verfügbar');
    }
  }
  loadUserStatistics(): void {
    const token = localStorage.getItem('token');
    const headers = new HttpHeaders().set('Authorization', `Bearer ${token}`);

    this.http.get<any>('https://api.educonnect-bmsd22a.bbzwinf.ch/progress/overview', { headers }).subscribe({
      next: (data) => {
        console.log('📊 Benutzer-Statistiken geladen:', data);
        this.statistics.completedCourses = data.completed_courses;
        this.statistics.completedExercises = data.completed_tasks;
      },
      error: (err) => {
        console.error('❌ Fehler beim Laden der Statistiken:', err);
      }
    });
  }


  formatDateToZurich(utcDateTime: string): string {
    if (!utcDateTime) return 'Keine Zeit verfügbar';
    const date = new Date(utcDateTime);
    const zurichTime = date.toLocaleString('de-CH', { timeZone: 'Europe/Zurich' });
    return zurichTime;
  }

}
