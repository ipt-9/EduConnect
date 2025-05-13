  import { Component, AfterViewInit } from '@angular/core';
  import {HttpClient, HttpClientModule, HttpHeaders} from '@angular/common/http';
  import {Router, RouterModule} from '@angular/router';
  import {CommonModule} from '@angular/common';

  declare var Stripe: any;

  @Component({
    selector: 'app-payment',
    templateUrl: './payment.component.html',
    standalone: true,
    imports: [CommonModule, RouterModule,HttpClientModule],
  })
  export class PaymentComponent implements AfterViewInit {

    stripe: any;

    constructor(private http: HttpClient, private router: Router) {}

    async ngAfterViewInit() {
      this.stripe = Stripe('pk_live_51R7Glr01266L6uW7fjUNnkgIQHjbj5SEdrATmp17J0TSnzeEzhbwrU4cURoHdpWpDi5Vp1dPVdGiYs1TH2K2Y6qm00ihGwqlCQ');

      const fetchClientSecret = async () => {
        const response = await fetch('https://api.educonnect-bmsd22a.bbzwinf.ch/create-checkout-session', {
          method: 'POST',
        });
        const { clientSecret } = await response.json();
        console.log(clientSecret);
        return clientSecret;
      };

      const checkout = await this.stripe.initEmbeddedCheckout({
        fetchClientSecret,
        onComplete: () => this.onPaymentSuccess() // 🆕 Nach Abschluss
      });

      checkout.mount('#checkout');
    }

    private onPaymentSuccess(): void {
      const token = localStorage.getItem('token');
      const headers = new HttpHeaders().set('Authorization', `Bearer ${token}`);

      this.http.post('https://api.educonnect-bmsd22a.bbzwinf.ch/activate-subscription', {}, { headers }).subscribe({
        next: () => {
          console.log('✅ Subscription aktiviert, weiterleiten zum Dashboard...');
          this.router.navigate(['/dashboard']);
        },
        error: (err) => {
          console.error('❌ Fehler beim Aktivieren der Subscription:', err);
        }
      });
    }
  }
