---
title: "Playing with fire -- an experiment with bots and public forms on websites"
date: 2024-09-21
tags: [hugo,open source,staticman]
---

In a previous post, I wrote about [Adding a Guestbook to Hugo website using Staticman](/posts/2023/06/guestbook-for-hugo-using-staticman/).  It was a fun weekend experiment, and I intentionally kept it "stupid simple" just like it used to be made in the old days -- fully knowing that it was not a good idea to expose the endpoint to the public.  I was curious to see how long it would take for the bots to find it and start spamming it.  In short, it took longer than expected, a real person found the website before the bots did, and none of the spam content submitted was ever published to the site because of the approval gateway / design for guestbook contributions.

The original post was posted June 2023, but the actual live [Guestbook](/guestbook/) page was published March 2023.  After a few months of no activity in [December 2023](https://github.com/bryankaraffa/b10a.co/pull/15) I decided to take the opportunity to be the first to sign the guestbook -- potentially entising other visitors to do the same.  

Sure enough -- in [February 2024](https://github.com/bryankaraffa/b10a.co/pull/16) a real person (I am assuming someone from Germany) signed the guestbook.  "Phil" left the message, `Das ist ziemlich cool` which I used Google to translate to "That's pretty cool".  This seemed genuine and so it was approved.  I had my first contribution -- and surpringly it was from a real person it seemed!  It was only [2 days after that](https://github.com/bryankaraffa/b10a.co/pull/18) when I got my first spam submission.  The spam was caught by the approval gateway and never published to the site.  I was happy to see that the design was working as intended.

With that said, I wanted to continue this experiment and there are plenty of solutions to address spam bots from even submitting content.  Most recently I made a change to the guestbook, [adding a "honeypot" strategy](https://github.com/bryankaraffa/b10a.co/commit/86b8d71e41a84f13b9784560821547d67ebe4763) to attempt to mitigate/prevent spam bots from even submitting content to the approval gateway.  I am curious to see how well this works, everything is live on the [Guestbook](/guestbook) as of this post.  

If you are a carbon-based lifeform, please feel free to sign the guestbook and let me know what you think of the site.  Happy web surfing all!
```