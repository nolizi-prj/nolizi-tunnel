# Acceptance cases · Spec 0001

Frozen at spec review, before implementation. The builder may not edit them;
a case that is wrong is fixed by amending the spec in the open and taking a
fresh cross-family spec review (CHARTER §3 requirement 2).

Every case below names **what execution makes it fail**. An assertion with no
such execution is decorative (L-006).

| # | Case | Go test | Fails when |
|---|---|---|---|
| **A-1** | A relay told nothing about TLS announces `http://<name>.<domain>` from `PublicURL`. | `core.TestPublicURLDefaultsToHTTP` | `PublicURL` hardcodes any scheme other than the configured one, or defaults to `https`. |
| **A-2** | A relay configured `https` announces `https://<name>.<domain>`. | `core.TestPublicURLHonoursConfiguredScheme` | The scheme is ignored — e.g. hardcoded `http` in place of the old hardcoded `https`. |
| **A-3** | `ParsePublicScheme` accepts `http`, `https`, `HTTPS`, ` https `, `https://`, and `""`→`http`; it refuses `ftp`, `htp`, `ws`, `HTTPs://x`, `http s`. | `core.TestParsePublicScheme` | An unknown scheme is coerced to a legal one instead of refused, or a legal spelling is refused. |
| **A-4** | A registry handed an unvalidated, illegal scheme announces `http`, never the illegal string. | `core.TestNewRegistryFailsClosedToHTTP` | `NewRegistry` interpolates whatever it was given. |
| **A-5** | The address the agent is handed at connect — surface 1, the CLI's first line — carries the relay's configured scheme, under both `http` (default) and `https`. | `relay.TestAuthResponseCarriesTheRelayScheme` | The auth response is built from a literal rather than from `PublicURL`. |
| **A-6** | `/_pumasi/status` — surface 2, the console — reports the same scheme, under both. | `relay.TestConsoleReportsTheRelayScheme` | The console keeps its own scheme, or the status view is built from a literal. |
| **A-7** | The zero-install ssh banner — surface 3, printed into the terminal of someone who installed nothing — carries the same scheme, under both. A real ssh client connects, opens a session channel, and reads the greeting. | `relay.TestSSHBannerCarriesTheRelayScheme` | The banner is built independently of `PublicURL`. |
| **A-8** | All three surfaces of **one** relay agree with each other and with `Registry.PublicURL`, under both schemes. This is the case a future copy-paste breaks. | `relay.TestAllThreeSurfacesAgreeOnTheScheme` | Any surface acquires its own scheme, however it is spelled. |
| **A-9** | A relay configured with a scheme it cannot honour does not start: `relay.New` returns `core.ErrUnknownScheme` and no relay. | `relay.TestRelayRefusesAnUnknownScheme` | An illegal scheme is silently coerced, which would restore the defect in a new spelling. |
| **A-10** | Configuring the scheme changes the scheme and nothing else: the host, the label's lowercasing, the base-domain normalisation and `TCPAddr` are unchanged under both. | `relay.TestSchemeChangesNothingButTheScheme` | The scheme value leaks into the host, or `TCPAddr` grows a scheme. |

## Not covered here, deliberately

- **That `https://` works.** No test in this repository can assert that,
  because whether TLS is terminated is a fact about the host in front of the
  relay. That is exactly why the scheme is configured rather than detected,
  and it is `BACKLOG.md` item 1 half (b).
- **The announce-before-bind race** (`BACKLOG.md` item 2). A-5 and A-7 read
  the auth response; they assert nothing about when the listener exists, and
  must not be read as covering it.
