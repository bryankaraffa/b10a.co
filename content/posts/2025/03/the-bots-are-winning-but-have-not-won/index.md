---
title: "The Bots Are Winning, But Have Not Won"
date: 2025-03-28
tags: [hugo,open source,staticman,spam]
---

In my previous post, ["Playing with Fire"](/posts/2024/09/playing-with-fire-an-experiment-with-bots/), I shared my experiment of adding a public [guestbook](/guestbook) to my Hugo site using Staticman. At the time, I knew it was only a matter of time before bots would find the form and start spamming it. Unfortunately my homebrew anti-spam measure did not work as intended, and there was a volume of spam being submitted that required additional measures to mitigate or prevent.

### The Current State of the Guestbook

Since launching the [Guestbook](/guestbook), I’ve seen a mix of genuine contributions (congrats, Phil for signing first!) and an steady number of spam submissions. Initially, I added a simple honeypot field to deter bots, but it wasn’t enough. The bots never really were deterred, and the spam submissions kept coming. While none of the spam made it to the live site due to the required approval gateway that was in the design, the anything that's not a genuine submission became a nuisance to manage.

### Fighting Back: Akismet and reCAPTCHA

To address the growing spam problem, I recently enabled two additional layers of protection to the Guestbook form:

1. **Akismet**: This service helps filter out spam by analyzing submissions and flagging suspicious ones. It’s widely used and integrates seamlessly with Staticman.
2. **reCAPTCHA**: By adding Google’s reCAPTCHA to the guestbook form, I’ve introduced a challenge that bots struggle to bypass. This should significantly reduce automated spam submissions.

Both measures are now live on the [Guestbook](/guestbook), and I’m optimistic that they’ll help keep the spam at bay while still allowing genuine visitors to leave their messages.

### Lessons Learned

This experiment has been a good learning lession in how sophisticated bots have become. While it’s frustrating to deal with spam, it’s also a reminder of the importance of designing systems that can adapt to evolving threats.

If you’re a "carbon-based lifeform" reading this, I’d love for you to test out the [guestbook](/guestbook) and let me know what you think of the site. Happy Surfing