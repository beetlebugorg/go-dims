# go-dims

## Commit messages

One short line, imperative mood, US plain language. State the change and nothing else.

- No history. Do not describe the previous behavior.
- No attribution. No sign-off, co-author, or tool references.
- No commentary. No reasoning or apologies.
- Add a body only when a commit carries several changes. Use plain statements, one per line.

```
Use HKDF-SHA256
Return an error when decryption fails
Remove leading v when building release tag
```

## Prose

For the README, `docs/`, comments, PR descriptions, and error messages.

- One idea per sentence. 20 words for instructions, 25 for descriptions.
- Active voice, present tense.
- One word, one meaning. No synonym for a term already used.
- No metaphors, idioms, or slang.
- Finite verbs, not gerund or participle phrases.
- Keep the articles. Three words maximum in a noun cluster.
- Positive statements. Instructions as commands.
- Six sentences per paragraph. Lists for sequences and sets.
- Keep technical names and technical verbs unchanged.

## Documentation

For the README and `docs/`. State what a thing is, what it does, and how to set it.

- No commentary. Do not argue for a setting or rank it against another.
- No history. Do not describe what the behavior was before.
- No rhetorical build-up. Do not withhold a term, restate a sentence for
  emphasis, or open with a question.
- Do not name a plan file, a review file, or a finding ID. This applies to
  commit messages, pull request descriptions, and code comments as well.

A setting page gives the name, one sentence on what it does, the default, the
rules that apply, and an example.

When you add or rename an environment variable, update the matching page in
`docs/docs/configuration/`. The name must match the code exactly.
