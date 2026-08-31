# Intent · The relay announces the scheme it is actually serving

**Published 2026-08-31 · `roadmap/BACKLOG.md` item 1, build half (a) only.**
CHARTER §2.1 gives an intent statement a 24-hour veto window. `roadmap/STAGE.md`
says `Alpha`, so CHARTER Part 0 applies: the window does not hold the work, and
a veto reverts rather than prevents. Recorded rather than pretended — this
statement was published at the same time as the work, not 24 hours ahead.

## The gap

Every tunnel this relay opens is handed an address that nothing answers.

Measured against the live relay at 15:30–15:37 UTC on 2026-08-31, not read off
the code:

```
$ curl -s http://pumasi.link/_pumasi/status
{"base_domain":"pumasi.link", ... "tunnels":[{"subdomain":"sshsteward",
 "url":"https://sshsteward.pumasi.link", ...}]}

$ curl https://sshsteward.pumasi.link/
curl: (7) Failed to connect to sshsteward.pumasi.link port 443 ... exit 7

$ : > /dev/tcp/pumasi.link/443   → Connection refused
$ : > /dev/tcp/pumasi.link/80    → connected
```

`Registry.PublicURL` (`core/route.go`) builds `"https://" + name + "." +
baseDomain` unconditionally. That string is the first line of output every user
of this product sees: the relay puts it in the auth response, the CLI prints
it, the ssh ingress banner prints it into the terminal of someone who installed
nothing, and the console renders it as a link. All four are wrong at once, and
they are wrong in the same way because they are all reading the same
unconditional constant.

The relay does not terminate TLS, and that is deliberate — the package header
of `cmd/pumasi-relay` says so, and it is the right call: an operator may want
ACME, a purchased certificate, or none at all on a private network. What is
wrong is not the choice. It is that the relay **states a consequence of a
choice it did not make and cannot observe.**

## What this intends

The scheme becomes a configured fact with a truthful default, decided in one
place and read by every surface that shows a person an address.

- A relay with nothing in front of it says `http://`, because that is what it
  serves.
- An operator who does put a TLS terminator in front says so, once, and every
  surface says `https://` together.
- A scheme the relay cannot honour is refused when it starts, not announced.

## What this deliberately does not intend

- **No TLS in the relay.** The seam stays exactly where it is. This change
  makes the relay honest about the seam; it does not move it.
- **Not the certificate.** Putting a wildcard certificate for `*.pumasi.link`
  in front of the relay on the Vultr host is `BACKLOG.md` item 1 half **(b)**,
  which that file marks *operator action, not a build*, and which
  `pumasi/DECISIONS.md` **Q-014** governs. Nothing here is a step toward
  taking it.
- **Not a deployment.** See the release note. The relay keeps no durable state
  and its one live tunnel carries this machine's own ssh access.
- **Not `BACKLOG.md` item 2** (the public TCP address is announced before
  anything listens on it). Same file, different root cause, its own review.

## How this could be wrong

The default flips what an existing operator's relay announces. A relay that
already has a terminator in front of it, upgraded without setting the new flag,
starts telling its users `http://` where `https://` worked. That is the real
harm this change can do, it is why the release note carries a veto window, and
the mitigation is that the flag is a single word in the unit file and the note
says so first rather than last.
