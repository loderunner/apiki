---
title: "apiki"
layout: hextra-home
---

<div class="hx:mt-6 hx:mb-6">
{{< hextra/hero-headline >}}
  Manage your environment&nbsp;<br class="sm:hx:block hx:hidden" />variables with ease
{{< /hextra/hero-headline >}}
</div>

<div>
{{< hextra/hero-subtitle >}}
  A clean, interactive TUI for organizing, selecting,&nbsp;<br class="sm:hx:block hx:hidden" />and applying environment variables across projects
{{< /hextra/hero-subtitle >}}
</div>

<div id="demo" class="hx:w-full" style="padding: 0 20%; margin: 4rem 0;">
</div>

<div class="hx:relative">
  <div class="hx:max-w-2xl hx:mx-auto hx:absolute hx:overflow-visible" style="position: absolute; top: -50px; left: 50px; z-index: 10; transform: rotate(-15deg); pointer-events: none;">
      <div style="background: linear-gradient(135deg, #ffd000 0%, #ffaa00 100%); clip-path: polygon(100% 50%, 81% 56%, 96% 69%, 77% 68%, 85% 85%, 68% 77%, 69% 96%, 56% 81%, 50% 100%, 43% 81%, 30% 96%, 31% 77%, 14% 85%, 22% 68%, 3% 69%, 18% 56%, 0% 50%, 18% 43%, 3% 30%, 22% 31%, 14% 14%, 31% 22%, 30% 3%, 43% 18%, 49% 0%, 56% 18%, 69% 3%, 68% 22%, 85% 14%, 77% 31%, 96% 30%, 81% 43%); width: 130px; height: 130px; display: flex; align-items: center; justify-content: center; filter: drop-shadow(3px 3px 8px rgba(0,0,0,0.25));">
        <span style="color:rgb(255, 255, 255); font-weight: 900; font-size: 14px; text-transform: uppercase; text-align: center; line-height: 1.2; padding: 20px; filter: drop-shadow(0 1px 0 rgba(0,0,0,0.25));">Install<br>now!</span>
      </div>
  </div>
</div>

<div class="hx:text-xl hx:text-center hx:pt-8" style="align-self: center;">

```shell
curl -fsSL https://github.com/loderunner/apiki/releases/latest/download/install.sh | sh
```

</div>

<div class="hx:mt-12"></div>

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Terminal-based Interface"
    subtitle="Navigate and manage variables with an interactive terminal UI and keyboard shortcuts. No more editing files by hand."
    class="hx:aspect-auto md:hx:aspect-[1.1/1] max-md:hx:min-h-[340px]"
    icon="terminal"
    style="background: radial-gradient(ellipse at 50% 80%,rgba(194,97,254,0.15),hsla(0,0%,100%,0));"
  >}}
  {{< hextra/feature-card
    title="Fuzzy Search"
    subtitle="Quickly find variables by name or description with built-in fuzzy search. Just start typing to filter your entries."
    class="hx:aspect-auto md:hx:aspect-[1.1/1] max-md:hx:min-h-[340px]"
    icon="search"
    style="background: radial-gradient(ellipse at 50% 80%,rgba(59,130,246,0.15),hsla(0,0%,100%,0));"
  >}}
  {{< hextra/feature-card
    title="Radio Groups"
    subtitle="Define multiple values for the same variable. Toggle between dev, staging, and production environments with a keypress."
    class="hx:aspect-auto md:hx:aspect-[1.1/1] max-md:hx:min-h-[340px]"
    icon="switch-horizontal"
    style="background: radial-gradient(ellipse at 50% 80%,rgba(16,185,129,0.15),hsla(0,0%,100%,0));"
  >}}
  {{< hextra/feature-card
    title=".env Integration"
    subtitle="Automatically discovers and loads .env files from your project directory. Your existing workflow, enhanced."
    class="hx:aspect-auto md:hx:aspect-[1.1/1] max-md:hx:min-h-[340px]"
    icon="document"
    style="background: radial-gradient(ellipse at 50% 80%,rgba(251,146,60,0.15),hsla(0,0%,100%,0));"
  >}}
  {{< hextra/feature-card
    title="Environment Import"
    subtitle="Import variables directly from your current shell environment. Quickly capture your existing setup with one command."
    class="hx:aspect-auto md:hx:aspect-[1.1/1] max-md:hx:min-h-[340px]"
    icon="download"
    style="background: radial-gradient(ellipse at 50% 80%,rgba(244,63,94,0.15),hsla(0,0%,100%,0));"
  >}}
  {{< hextra/feature-card
    title="Shell Integration"
    subtitle="Works seamlessly with bash, zsh, and fish. Just eval the output and your environment variables are set."
    class="hx:aspect-auto md:hx:aspect-[1.1/1] max-md:hx:min-h-[340px]"
    icon="code"
    style="background: radial-gradient(ellipse at 50% 80%,rgba(6,182,212,0.15),hsla(0,0%,100%,0));"
  >}}
{{< /hextra/feature-grid >}}

<script src="/apiki/asciinema/asciinema-player.min.js"></script>
  <script>
    AsciinemaPlayer.create('full_demo.cast', document.getElementById('demo'), {
      autoplay: true,
      loop: true,
      rows: 18,
      poster: 'npt:0',
      speed: 2,
      idleTimeLimit: 2 });
  </script>