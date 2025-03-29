import { Routes } from '@angular/router';
import {LoginComponent} from './pages/login/login.component';
import {ContactComponent} from './pages/contact/contact.component';
import {ErrorpageComponent} from './pages/errorpage/errorpage.component';
import {HomepageComponent} from './pages/homepage/homepage.component';
import {RegisterComponent} from './pages/register/register.component';

export const routes: Routes = [
  { path: 'homepage', component: HomepageComponent },
  { path: 'login', component: LoginComponent },
  { path: 'register', component: RegisterComponent },
  { path: 'contact', component: ContactComponent },
  { path: 'errorpage', component: ErrorpageComponent },
];
