![6aaabeab-4079-4277-9213-9e677194a206](https://github.com/MlDenis/prometheus_wannabe/assets/89723953/73238174-550f-4dc1-b637-c167179fde07)

# prometheus_wannabe

Repository for the track "Server for collecting metrics and alerting".

## Template update

To be able to receive updates to autotests and other parts of the template, run the command:
```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

To update the autotest code, run the command:

```
git fetch template && git checkout template/main .github
```

Then add the resulting changes to your repository.

## Running Autotests

To successfully run autotests, name the branches `iter<number>`, where `<number>` is the serial number of the increment. For example, a branch named `iter4` will run autotests for increments one through four.

When merging a branch with an increment to the `main` branch, all autotests will be run.

Read more about local and automatic launch in the [README of autotests](https://github.com/Yandex-Practicum/go-autotests).