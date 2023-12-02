---
title: "Adding a Guestbook to Hugo website using Staticman"
date: 2023-06-01
tags: [hugo,staticman,open source,free]
---
This website is powered by Hugo which is a static site generator, and static websites typically do not support user-generated content like forms (which require server-side handling).    In some cases when someone visits a website, they may also want to acknowledge their visit.  I decided to build a Guestbook to explore adding user-generated content to this otherwise static website.

A Guestbook allows visitors to leave details such as their name, and an optional comment like where they are from or how they found the site.

## Deploy Staticman
In strongly recommend following the official [Staticman Docs > Getting Started](https://staticman.net/docs/getting-started.html) page for details on how to deploy Staticman.  *I did not do that in this case..*

As of today Staticman team recommends deploying to Heroku (which has a free tier). Unfortunately, I do not have a Heroku account today and they require a Credit Card for that free tier.

In an effort to avoid giving another Cloud Vendor my payment details, I made the decision to use my existing Google Cloud Platform account to run Staticman on [Google Cloud Run](https://cloud.google.com/run/).  Google Cloud Run is currently offering a free tier of 2 Million requests per month.  I am not concerned right now about using more than that between all my Cloud Run deployments combined.

To run Staticman in Google Cloud Run we need to build and publish the Docker image to a Docker Image registry.  Google Cloud Run currently supports pulling [Public container images from Google Cloud Artifact Registry, Google Cloud Container Registry, or Docker Hub.
](https://cloud.google.com/run/docs/deploying#images).  This makes deploying to Google Cloud run slightly more difficult because the Staticman maintainers do not publish official Docker images to Docker Hub.  I submitted [eduardoboucas/staticman #457](https://github.com/eduardoboucas/staticman/pull/457) which would add the workflows to build & publish their image to GitHub Packages Container Registry.

### Build & Publish Staticman image to Docker Hub
Ideally the Staticman team [would provide an official Docker image from their namespace](https://github.com/eduardoboucas/staticman/pull/457), but in the meantime I put together a GitHub Actions Workflow to build and publish the Docker image using Staticman source.

> GitHub Actions Workflow: [docker-build.yml](https://github.com/bryankaraffa/staticman/blob/master/.github/workflows/docker-build.yml)

> [Docker Hub Image](https://hub.docker.com/r/bryankaraffa/staticman)
>
> `docker pull bryankaraffa/staticman:master`

GitHub Actions, GitHub Container Registry, and Docker Hub are free for public repositories/projects.  All of this is zero-cost and actually does not even require a credit card to implement.

### Deploy Staticman to Google Cloud Run
With the Docker Image for Staticman available from [bryankaraffa/staticman](https://hub.docker.com/r/bryankaraffa/staticman) for download from Docker Hub, I can now create a deployment in Google Cloud Run.  This can be done via the UI, or via CLI:
```sh
gcloud run deploy staticman \
--image=bryankaraffa/staticman \
--timeout=30 \
--max-instances=10 \
--cpu-boost \
--region=us-central1 \
--project=__GCP_PROJECT_NAME__ \
 && gcloud run services update-traffic staticman --to-latest
```

After deployment, I can use the Service URL provided for my form.  To test Staticman is listening for requests as expected we can make a simple request to the Service URL
```sh
$ curl -i https://staticman-4labc7defg-ij.a.run.app/
HTTP/2 200
x-powered-by: Express
access-control-allow-origin: *
access-control-allow-headers: Origin, X-Requested-With, Content-Type, Accept
content-type: text/html; charset=utf-8
etag: W/"23-aOCwT7bnC/aXwxLQsIsRLd3YDxk"
x-cloud-trace-context: f778c500503235405b3db136e390b842;o=1
date: Sun, 05 Mar 2023 21:49:05 GMT
server: Google Frontend
content-length: 35
alt-svc: h3=":443"; ma=2592000,h3-29=":443"; ma=2592000,h3-Q050=":443"; ma=2592000,h3-Q046=":443"; ma=2592000,h3-Q043=":443"; ma=2592000,quic=":443"; ma=2592000; v="46,43"

Hello from Staticman version 3.0.0!
```

## Staticman Site Configuration `staticman.yml`
When a request is made to the Staticman deployment (i.e. from a form on our site) -- Staticman will look for a `staticman.yml` file in the root of the repository, where various configuration parameters will be defined (like `allowedFields`,`requiredFields`, etc.. see [Staticman Docs > Site configuration file](https://staticman.net/docs/configuration))

Here is the resulting block in `staticman.yml` to handle a simple Guestbook form:
```yml
guestbook:
  allowedFields: ["name", "message"]
  allowedOrigins: ["b10a.co"]
  branch: "main"
  commitMessage: "New Guestbook Post from {fields.name}"
  filename: "entry{@timestamp}"
  format: "yaml"
  generatedFields:
    date:
      type: date
      options:
        format: "timestamp-seconds"
  moderation: true
  name: "b10a.co"
  path: "data/guestbook"
  requiredFields: ["name"]
```
Some of the basic configurations are:
  - 2 fields: `name`, and `message`
  - `name` is required
  - All submissions will require approval (Pull Request) before getting published
  - Data files output by Staticman will output to `data/guestbook/entries<timestamp>.yml

The full Staticman Site configuration for this site is public here: [`staticman.yml`](https://github.com/bryankaraffa/b10a.co/blob/main/staticman.yml)


## Add Guestbook Shortcodes and Page to Hugo Site
Now that we have a Staticman deployment configured and ready to handle requests, we need to add a Guestbook page with the form and previous guestbook entries.

Hugo has a concept of [Shortcodes](https://gohugo.io/content-management/shortcodes/) which are simple snippets inside your content files calling built-in or custom templates.  We can use shortcodes to dynamically format the Guestbook form and Guestbook entries.

`layouts/shortcodes/guestbook-form.html`
{{<sup>}}[[View on GitHub](https://github.com/bryankaraffa/b10a.co/blob/main/layouts/shortcodes/guestbook-form.html)]{{</sup>}}
```html
<form method="POST" action="https://staticman-4loez7ackq-uc.a.run.app/v3/entry/github/bryankaraffa/b10a.co/main/guestbook">
  <input name="options[redirect]" type="hidden" value="{{ absURL "/guestbook" }}?success=true">
  <label>Name or Callsign <input name="fields[name]" type="text" alt="Name or Callsign" placeholder="Joe Bob"></label><br/>
  <label>Optional Message <br/><textarea name="fields[message]" alt="Message" style="width: 90%; min-width: 100px;" placeholder="Hello!"></textarea></label><br/>

  <button type="submit">Submit</button>
</form>
<small><i>
  All messages submitted are moderated, and any extra HTML/Markdown will be stripped -- only plaintext messages will be approved.
</i></small><br />
```

`layouts/shortcodes/guestbook-entries.html`
{{<sup>}}[[View on GitHub](https://github.com/bryankaraffa/b10a.co/blob/main/layouts/shortcodes/guestbook-entries.html)]{{</sup>}}
```html
<section class="guestbook" id="guestbook-section">
  <!-- Existing guestbook -->
  <small><i>Guestbook entries are sorted by date, most recent at top to oldest at the bottom</i></small>
  <div class="guestbook__existing">
  {{ if .Site.Data.guestbook }}
    {{ range sort .Site.Data.guestbook "date" "desc"  }}
    <blockquote>
      <!-- htmlEscape the content just in case there are attempted injections -->
      <span class="post-header">{{ safeHTML .message }}</span><br/>
      <!-- Entry metadata with date formatted as "January 6, 2006" -->
      <span class="post-meta">- <b>{{ safeHTML .name }}</b> on {{ time.Format "January 2, 2006" .date }}</span><br/>
    </blockquote>
    {{ end }}
    {{ else }}
    <blockquote>
      No guestbook entries yet -- you could be the first!
    </blockquote>
    {{ end }}
  </div>
</section>
```

We also need to add a new page to our site, which will use these `{{</* guestbook-form */>}}` and `{{</* guestbook-entries */>}}` shortcodes.

`content/guestbook.md`
{{<sup>}}[[View on GitHub](https://github.com/bryankaraffa/b10a.co/blob/main/content/guestbook.md)]{{</sup>}}
```html
---
title: "Guestbook"
date: 2023-03-01
draft: false
type: page
nodate: true
hidemeta: true
---
If you are interested in saying "Hello" or leaving a message -- please [sign the guestbook](#sign-the-guestbook)!

## Sign the Guestbook

{{</* guestbook-form slug={{page.title}} */>}}

## Guestbook Entries

{{</* guestbook-entries */>}}
```

## The Result
Putting it all together, this is what it looks like
{{< figure src="Screenshot 2023-03-05 140426.png" >}}

You can see it working live on the [Guestbook](/guestbook) Page.

When a visitor submits the Guestbook Form, Staticman handles the request and will create a Pull Request for approval with the details.  At this point if I approve and merge the PR, the submission will go through the normal Hugo page publishing pipeline and get published to the Guestbook.

## Afterthoughts

This was a simple weekend project which helped me explore Staticman, open up the possibility for additional user-generated content on the website.

I also feel it brought a bit of nostalgia. When I first started with computers, the internet was young and Guestbooks were *very* popular -- along with animated "Under Construction" images, blinking text, and other gems of the era.  This guestbook is much more simple compared to what we built back then, and I am curious to see if Guestbooks still have any place on the internet in 2023 and beyond..
