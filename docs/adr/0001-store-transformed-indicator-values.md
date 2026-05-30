# Store transformed indicator values at collection time

The `transform.expr` step applies simple math (e.g. `vars.balance / 100`) to extracted
vars during a run, and the **transformed value is what gets persisted** as the DataPoint.
We chose this over a transform-at-read model (store raw, apply a factor only when
displaying/serving) so that the DataPoint stays the single source of truth: Rules,
history, the dashboard, and the public API all see the same scaled number with no
special-casing.

The trade-off: historical DataPoints are baked at the value collected under the
definition in force at the time. Changing an expression later does not rewrite old
samples, so a factor change produces an honest discontinuity in the time series rather
than retroactively re-scaling history.

The step reuses the existing goja sandbox (as `script.js.*` does) rather than adding a
dedicated expression library; `transform.expr` is a narrower, sugar-over-script
convenience that only exposes `vars` and requires a finite numeric result.
