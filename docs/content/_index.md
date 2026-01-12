---
title: "apiki"
layout: hextra-home
---

{{< hextra/hero-badge >}}
  <div class="hx:w-2 hx:h-2 hx:rounded-full hx:bg-primary-400"></div>
  <span>Terminal-based</span>
  {{< icon name="terminal" attributes="height=14" >}}
{{< /hextra/hero-badge >}}

<div class="hx:mt-6 hx:mb-6">
{{< hextra/hero-headline >}}
  Manage your environment&nbsp;<br class="sm:hx:block hx:hidden" />variables with ease
{{< /hextra/hero-headline >}}
</div>

<div class="hx:mb-12">
{{< hextra/hero-subtitle >}}
  A clean, interactive TUI for organizing, selecting,&nbsp;<br class="sm:hx:block hx:hidden" />and applying environment variables across projects
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx:mb-6">
{{< hextra/hero-button text="Get Started" link="docs/getting-started" >}}
</div>

<div class="hx:mt-12"></div>

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Visual Interface"
    subtitle="Navigate and manage variables with an interactive terminal UI and keyboard shortcuts. No more editing files by hand."
    class="hx:aspect-auto md:hx:aspect-[1.1/1] max-md:hx:min-h-[340px]"
    icon="cursor-click"
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
  >}}
  {{< hextra/feature-card
    title="Environment Import"
    subtitle="Import variables directly from your current shell environment. Quickly capture your existing setup with one command."
    class="hx:aspect-auto md:hx:aspect-[1.1/1] max-md:hx:min-h-[340px]"
  >}}
  {{< hextra/feature-card
    title="Shell Integration"
    subtitle="Works seamlessly with bash, zsh, and fish. Just eval the output and your environment variables are set."
    class="hx:aspect-auto md:hx:aspect-[1.1/1] max-md:hx:min-h-[340px]"
  >}}
{{< /hextra/feature-grid >}}
