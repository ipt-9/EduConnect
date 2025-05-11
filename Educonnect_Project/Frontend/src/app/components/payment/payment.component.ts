import { Component, AfterViewInit } from '@angular/core';

declare var Stripe: any; // falls kein TypeScript-Typ installiert

@Component({
  selector: 'app-payment',
  templateUrl: './payment.component.html',
})
export class PaymentComponent implements AfterViewInit {

  stripe: any;

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
    });

    checkout.mount('#checkout');
  }
}
