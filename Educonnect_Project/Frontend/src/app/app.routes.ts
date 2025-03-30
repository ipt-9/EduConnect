import { Routes } from '@angular/router';
import {LoginComponent} from './pages/login/login.component';
import {ContactComponent} from './pages/contact/contact.component';
import {ErrorpageComponent} from './pages/errorpage/errorpage.component';
import {HomepageComponent} from './pages/homepage/homepage.component';
import {RegisterComponent} from './pages/register/register.component';
import {CodingSpaceComponent} from './components/coding-space/coding-space.component';
import {CourseListComponent} from './components/course-list/course-list.component';
import {TasksListComponent} from './components/tasks-list/tasks-list.component';
import {AuthGuard} from './auth.guard';

export const routes: Routes = [
  { path: 'homepage', component: HomepageComponent, canActivate: [AuthGuard] },
  { path: 'codingSpace', component: CodingSpaceComponent, canActivate: [AuthGuard] },
  { path: 'login', component: LoginComponent },
  { path: 'register', component: RegisterComponent },
  { path: 'contact', component: ContactComponent, canActivate: [AuthGuard] },
  { path: 'errorpage', component: ErrorpageComponent, canActivate: [AuthGuard] },
    { path: 'courselist', component: CourseListComponent, canActivate: [AuthGuard] },
    { path: 'taskslist', component: TasksListComponent, canActivate: [AuthGuard] }
];
