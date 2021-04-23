---
title: "Contribute"
date: 2017-08-30T19:07:12-03:00
weight: 16
menu: main
---

Did you found an bug or you would like to suggest a new feature? I'm open for feedback. Please open a new [issue](https://github.com/getupio-undistro/undistro/issues) and let me know what you think.

You're also welcome to contribute with [pull requests](https://github.com/getupio-undistro/undistro/pulls).

## Contribution Guidelines

### Introduction

This document explains how to contribute changes to the UnDistro project.

### Bug reports

Please search the issues on the issue tracker with a variety of keywords to ensure your bug is not already reported.

If unique, [open an issue](https://github.com/getupio-undistro/undistro/issues/new) and answer the questions so we can understand and reproduce the problematic behavior.

To show us that the issue you are having is in UnDistro itself, please write clear, concise instructions so we can reproduce the behavior (even if it seems obvious). The more detailed and specific you are, the faster we can fix the issue. Check out [How to Report Bugs Effectively](http://www.chiark.greenend.org.uk/~sgtatham/bugs.html).

Please be kind, remember that UnDistro comes at no cost to you, and you're getting free help.

### Discuss your design

The project welcomes submissions but please let everyone know what you're working on if you want to change or add something to the UnDistro repository.

Before starting to write something new for the UnDistro project, please [open discussion here](https://github.com/getupio-undistro/undistro/discussions/new).

This process gives everyone a chance to validate the design, helps prevent duplication of effort, and ensures that the idea fits inside the goals for the project and tools. It also checks that the design is sound before code is written; the code review tool is not the place for high-level discussions.

### Code review

Changes to UnDistro must be reviewed before they are accepted, no matter who makes the change even if it is an owner or a maintainer.

Please try to make your pull request easy to review for us. Please read the "[How to get faster PR reviews](https://github.com/kubernetes/community/blob/main/contributors/devel/faster_reviews.md)" guide, it has lots of useful tips for any project you may want to contribute. Some of the key points:

* Make small pull requests. The smaller, the faster to review and the more likely it will be merged soon.
* Don't make changes unrelated to your PR. Maybe there are typos on some comments, maybe refactoring would be welcome on a function... but if that is not related to your PR, please make *another* PR for that.
* Split big pull requests into multiple small ones. An incremental change will be faster to review than a huge PR.

### Sign your work

The sign-off is a simple line at the end of the explanation for the patch. Your signature certifies that you wrote the patch or otherwise have the right to pass it on as an open-source patch.


Please use your real name, we really dislike pseudonyms or anonymous contributions. We are in the open-source world without secrets. If you set your `user.name` and `user.email` git configs, you can sign your commit automatically with `git commit -s`.

## Running tests

Clone the repository and run tests.

```
make test
```
