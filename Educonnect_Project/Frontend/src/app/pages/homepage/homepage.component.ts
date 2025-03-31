import {
  Component,
  AfterViewInit,
  Renderer2,
  Inject,
  ElementRef
} from '@angular/core';
import Chart from 'chart.js/auto';
import {DOCUMENT} from '@angular/common';
import {Router, RouterLink} from '@angular/router';
import {HttpClient} from '@angular/common/http';

@Component({
  selector: 'app-homepage',
  templateUrl: './homepage.component.html',
  imports: [
    RouterLink
  ],
  styleUrls: ['./homepage.component.scss']
})
export class HomepageComponent implements AfterViewInit {
  constructor(
    private renderer: Renderer2,
    private router: Router,
    @Inject(DOCUMENT) private document: Document
  ) {}

LoginBtn () {
    this.router.navigate(['/login']);
}

RegisterBtn () {
    this.router.navigate(['/register']);
}

  currentStep = 1;
  moritzSpeechIndex = 1;
  fynnSpeechIndex = 0;
  danSpeechIndex = 0;
  tomasSpeechIndex = 0;
  levinSpeechIndex = 0;

  moritzSpeeches = [
    `Hi, I’m Moritz 👨‍💻<br>I’ve helped shape this platform from the ground up — let me show you how it works.`,
    `This section lets you <strong>code directly</strong> in your browser — no setup needed! 🧑‍💻✨`,
    `You’ll get real-time feedback, security, and a smooth dev experience.`,
    `It's the perfect place to experiment and build your projects. Let’s continue! 🚀`
  ];

  fynnSpeeches = [
    `Yo! I'm Fynn 👋<br>Here to show you how we code, collab, and conquer together!`,
    `This part is all about <strong>learning together</strong>. You can join live sessions, share code, and vote on ideas. 🤝`,
    `We designed this for teamwork — feedback loops, project threads, and mentor chats.`,
    `Ready to grow with others? Let’s go! 💪`
  ];

  danSpeeches = [
    `Yo! I'm Dan — your frontend guy.<br><br>Let's talk <strong>feedback</strong>. 💬`,
    `This section is where students share suggestions, praise, or ideas.<br><br>It's like our community board.`,
    `You can view discussions, post your thoughts, or vote on feedback.`,
    `Pretty cool, right? Collaboration is how we grow.<br><br>Let’s keep going! 🚀`
  ];

  tomasSpeeches = [
    `Hey! I’m Tomas, your backend buddy 🔧<br><br>Let’s explore your progress!`,
    `This graph shows your weekly learning — HTML, CSS, JavaScript and more! 📈`,
    `On the right, you'll see a breakdown of the languages you’ve practiced.`,
    `It’s all about balance — track your growth and aim for consistency! 💡`,
    `That’s it from me. Keep pushing forward, one line of code at a time! 👊`
  ];

  levinSpeeches = [
    `Hey! I’m Levin — I help keep things running smoothly behind the scenes. 🛠️<br><br>Let’s explore your pricing options!`,
    `Here you’ll find three plans: Basic, Premium, and Business.`,
    `Each plan is designed for different levels — whether you're solo, premium, or part of a school.`,
    `Pick the one that fits your goals best — and yes, no ads in Premium 😎`,
    `That’s a wrap! Time to choose your coding journey. 🧑‍🚀`
  ];

  ngAfterViewInit(): void {
    const onboardingModal = this.document.getElementById('onboardingModal');
    const tourOverlay = this.document.getElementById('tourOverlay');
    const hasSeenTour = localStorage.getItem('eduTourDone');

    if (!hasSeenTour && onboardingModal) {
      onboardingModal.style.display = 'flex';
      document.body.style.overflow = 'hidden';
    } else if (tourOverlay) {
      tourOverlay.style.display = 'none';
      document.body.style.overflow = '';
    }

    this.setupScrollEffects();
    this.setupFadeObserver();
    this.setupBarObserver();
    this.setupChartObserver();
  }

  startTour(): void {
    localStorage.setItem('eduTourDone', 'true');
    this.closeModal();

    const featureSection = this.document.querySelector('#Features');
    featureSection?.scrollIntoView({ behavior: 'smooth' });

    setTimeout(() => {
      this.showElementById('tourOverlay');
      this.showElementById('guideStep1');
      featureSection?.classList.add('highlight');
      this.document.body.classList.add('tour-active');

      const guideSpeech = this.document.getElementById('guideSpeech');
      if (guideSpeech) guideSpeech.innerHTML = this.moritzSpeeches[0];
    }, 800);
    setTimeout(() => {
      document.body.style.overflow = 'hidden';
    },1000)

  }

  nextGuideStep(): void {
    const el = (id: string) => this.document.getElementById(id);

    const featureSection = el('Features');
    const collabSection = this.document.querySelector('.features.reverse') as HTMLElement;
    const feedbackSection = el('Feedback');
    const progressSection = el('ProgressCharts');
    const pricingSection = el('Pricing');

    const moritzSpeech = el('guideSpeech');
    const fynnSpeech = el('guideSpeechFynn');
    const danSpeech = el('guideSpeechDan');
    const tomasSpeech = el('guideSpeechTomas');
    const levinSpeech = el('guideSpeechLevin');

    switch (this.currentStep) {
      case 1:
        moritzSpeech!.innerHTML = this.moritzSpeeches[this.moritzSpeechIndex];
        this.currentStep++;
        break;

      case 2:
        this.moritzSpeechIndex++;
        if (this.moritzSpeechIndex < this.moritzSpeeches.length) {
          moritzSpeech!.innerHTML = this.moritzSpeeches[this.moritzSpeechIndex];
        } else {
          this.hideElementById('guideStep1');
          featureSection?.classList.remove('highlight');
          this.showElementById('guideStep2');
          collabSection?.scrollIntoView({ behavior: 'smooth' });
          collabSection?.classList.add('highlight');
          fynnSpeech!.innerHTML = this.fynnSpeeches[this.fynnSpeechIndex];
          this.currentStep++;
        }
        break;

      case 3:
        this.fynnSpeechIndex++;
        if (this.fynnSpeechIndex < this.fynnSpeeches.length) {
          fynnSpeech!.innerHTML = this.fynnSpeeches[this.fynnSpeechIndex];
        } else {
          this.hideElementById('guideStep2');
          collabSection?.classList.remove('highlight');
          this.showElementById('guideStep3');
          feedbackSection?.scrollIntoView({ behavior: 'smooth' });
          feedbackSection?.classList.add('highlight');
          danSpeech!.innerHTML = this.danSpeeches[this.danSpeechIndex];
          this.currentStep++;
        }
        break;

      case 4:
        this.danSpeechIndex++;
        if (this.danSpeechIndex < this.danSpeeches.length) {
          danSpeech!.innerHTML = this.danSpeeches[this.danSpeechIndex];
        } else {
          this.hideElementById('guideStep3');
          feedbackSection?.classList.remove('highlight');
          this.showElementById('guideStep4');
          progressSection?.scrollIntoView({ behavior: 'smooth' });
          progressSection?.classList.add('highlight');
          tomasSpeech!.innerHTML = this.tomasSpeeches[this.tomasSpeechIndex];
          this.currentStep++;
        }
        break;

      case 5:
        this.tomasSpeechIndex++;
        if (this.tomasSpeechIndex < this.tomasSpeeches.length) {
          tomasSpeech!.innerHTML = this.tomasSpeeches[this.tomasSpeechIndex];
        } else {
          this.hideElementById('guideStep4');
          progressSection?.classList.remove('highlight');
          this.showElementById('guideStep5');
          pricingSection?.scrollIntoView({ behavior: 'smooth' });
          pricingSection?.classList.add('highlight');
          levinSpeech!.innerHTML = this.levinSpeeches[this.levinSpeechIndex];
          this.currentStep++;
        }
        break;

      case 6:
        this.levinSpeechIndex++;
        if (this.levinSpeechIndex < this.levinSpeeches.length) {
          levinSpeech!.innerHTML = this.levinSpeeches[this.levinSpeechIndex];
        } else {
          this.hideElementById('guideStep5');
          pricingSection?.classList.remove('highlight');
          window.scrollTo({ top: 0, behavior: 'smooth' });
          this.hideElementById('tourOverlay');
          this.document.body.classList.remove('tour-active');
          this.resetTour();
          document.body.style.overflow = '';
        }
        break;
    }
  }

  closeModal(): void {
    this.hideElementById('onboardingModal');
  }

  private showElementById(id: string): void {
    const el = this.document.getElementById(id);
    if (el) el.style.display = 'flex';
  }

  private hideElementById(id: string): void {
    const el = this.document.getElementById(id);
    if (el) el.style.display = 'none';
  }

  private resetTour(): void {
    this.currentStep = 1;
    this.moritzSpeechIndex = 0;
    this.fynnSpeechIndex = 0;
    this.danSpeechIndex = 0;
    this.tomasSpeechIndex = 0;
    this.levinSpeechIndex = 0;
  }

  private setupScrollEffects(): void {
    window.addEventListener('scroll', () => {
      const navbar = this.document.querySelector('.navbar');
      if (navbar) navbar.classList.toggle('scrolled', window.scrollY > 50);

      this.document.querySelectorAll('[data-speed]').forEach((el) => {
        const speed = parseFloat(el.getAttribute('data-speed') || '1');
        (el as HTMLElement).style.transform = `translateY(${window.scrollY * speed}px)`;
      });
    });
  }

  private setupFadeObserver(): void {
    const fadeInElements = this.document.querySelectorAll('.fade-in, .child-fade');

    const fadeObserver = new IntersectionObserver((entries, observer) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add('visible');
          const children = entry.target.querySelectorAll('.child-fade');
          children.forEach((child, i) => {
            (child as HTMLElement).style.transitionDelay = `${i * 0.2}s`;
            child.classList.add('visible');
          });
          observer.unobserve(entry.target);
        }
      });
    }, { threshold: 0.2 });

    fadeInElements.forEach((el) => fadeObserver.observe(el));
  }

  private setupBarObserver(): void {
    const barGraphs = this.document.querySelectorAll('.bar-graph');

    // Wait for next animation frame to ensure styles/rendering are complete
    requestAnimationFrame(() => {
      const barObserver = new IntersectionObserver((entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const bars = entry.target.querySelectorAll('.bar');
            bars.forEach((bar: any) => {
              const height = bar.dataset.value;
              bar.style.height = height;
            });
          }
        });
      }, { threshold: 0.5 });

      barGraphs.forEach((graph) => barObserver.observe(graph));
    });
  }


  private setupChartObserver(): void {
    const chartContainer = this.document.querySelector('.language-pie-wrapper');
    if (!chartContainer) return;

    const chartObserver = new IntersectionObserver((entries, observer) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          this.renderDoughnutChart();
          observer.unobserve(entry.target);
        }
      });
    }, { threshold: 0.3 });

    chartObserver.observe(chartContainer);
  }

  private renderDoughnutChart(): void {
    const canvas = this.document.getElementById('languagePieChart') as HTMLCanvasElement;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    new Chart(ctx, {
      type: 'pie',
      data: {
        labels: ['Python', 'JavaScript', 'C#', 'GoLang'],
        datasets: [{
          data: [40, 30, 20, 10],
          backgroundColor: ['#f1562f', '#ff8e42', '#ff9a76', '#121212'],
          borderColor: '#fff',
          borderWidth: 2
        }]
      },
      options: {
        responsive: false,
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              label: function (ctx) {
                return `${ctx.label}: ${ctx.parsed}%`;
              }
            }
          }
        }
      }
    });
  }
}
