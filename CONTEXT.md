# SiphonGear

A configuration-driven collection and metrics platform: scheduled pipelines fetch data from websites, extract indicators, and persist them as time series.

## Language

**Site**:
A monitored website. Carries a set of tags used to group and target it.
_Avoid_: Target (in the website sense), host, endpoint

**Collector**:
A collection task bound to one Site, defined as a pipeline of steps.
_Avoid_: Job, crawler, task

**Indicator**:
A named value a Collector extracts from a run, tracked over time.
_Avoid_: Metric, field, datapoint (a DataPoint is a single recorded sample of an Indicator)

**Rule**:
A threshold rule that evaluates a condition over an Indicator to flag it on the dashboard and fire notifications.
_Avoid_: Alert (an alert is the outcome of a Rule firing), threshold

**Target**:
The selection of which Sites a Rule applies to. Composed of an include side (`all` Sites, or only Sites whose tags intersect the Rule's include tags) and an exclude side (Sites carrying any Exclude Tag are dropped). Exclusion always wins and applies to every target type.
_Avoid_: Scope, filter

**Exclude Tags**:
Site tags that remove a Site from a Rule's Target even when the include side would have selected it. Sites with no tags are never excluded.
_Avoid_: Blocklist, deny tags

**Tag**:
A free-form label on a Site, used by Rule Targets for include and exclude matching.
