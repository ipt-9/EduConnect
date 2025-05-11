import { Component, AfterViewInit } from '@angular/core';

declare var Stripe: any; // falls kein TypeScript-Typ installiert

@Component({
  selector: 'app-payment',
  templateUrl: './payment.component.html',
})
export class PaymentComponent implements AfterViewInit {

  stripe: any;

  async ngAfterViewInit() {
    this.stripe = Stripe('pk_test_51R7Gm402US8DmCztGSk8YNd1PFN6rt7KmIA7BJl1ZrfYKXRptltbvshrxmsivJfUeccG9qs7VGMMqZ3Yye41cO0o00tQ39OB61');

    const fetchClientSecret = async () => {
      const response = await fetch('api.educonnect-bmsd22a.bbzwinf.ch/create-checkout-session', {
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
