---
date: 2016-11-13T18:02:45Z
title: Extending Ladder
menu:
  main:
    parent: Operating
    weight: 34
---

Sometimes we want to scale our custom targets, or apply custom logic on filters or arrengers,
or our company uses a very strange metric system where all our metrics are, in order you can
solve this *problems* Ladder lets you extend using [Go plugins](https://golang.org/pkg/plugin).

Extending Ladder to add custom logic is very easy!, as was described in the [blocks] section, Ladder
is made up of 5 type of blocks: gatherers, arrangers, solvers, filters and scalers. You can create any
custom kind of this blocks but it has some requirements.

## Requirements

* You need to compile your plugin against the Ladder version that you want to use
* It will use Go >=1.8
* Requires to autoregister the plugin (See the example)

## Blocks

Depending on each block we want to implement you will need to satisfy specific interfaces. Lets start

### Gatherer

```golang
type Gatherer interface {
	Gather(ctx context.Context) (types.Quantity, error)
}
```

* `Gather` method receives a context an returns a quantity and an error, this method should get the 
    metrics from the external source and return them

### Arranger

```golang
type Arranger interface {
	Arrange(ctx context.Context, inputQ, currentQ types.Quantity) (newQ types.Quantity, err error)
}
```

* `Arrange` method receives a context and 2 quantities, the first quantity is the value obtained from the 
    gatherer and the second one is the current scaling target quantity, it should return the wanted
    quantity to set up on the scaling target (finters and solvers may change this quantity before reaching the scaler) and an error if required

### Solver

```golang
type Solver interface {
	Solve(ctx context.Context, qs []types.Quantity) (types.Quantity, error)
}
```

* `Solve` methods receives a context and an slice of quantities, this values will be all the quantities
    got from all the inputters (gatherer + arranger) it should return one, and an error if required

### Filter

```golang
type Filterer interface {
	Filter(ctx context.Context, currentQ, newQ types.Quantity) (q types.Quantity, br bool, err error)
}
```

* `Filter` method receives a context and 2 quantities, the first one is the currenr scalign target quantity
    and the second one is the new quantity (from the solver or previous filter). It returns a quantity
    a boolean that if tru it will break teh chain and an error.

### Scaler

```golang
type Scaler interface {
	Current(ctx context.Context) (types.Quantity, error)
	Scale(ctx context.Context, newQ types.Quantity) (scaledQ types.Quantity, mode types.ScalingMode, err error)
	Wait(ctx context.Context, scaledQ types.Quantity, mode types.ScalingMode) error
}
```

* `Current` method receives a context and returns a quantity that should be the current quantity an scaling target has, and an error if required
* `Scale` receives the context and the scaling quantity decided, it should scale the scaling target and return the quantity scalated to, the scaling mode (ScalingUp, ScalingDown, NotScaling) and an error if required
* `Wait` method receives a context, the scaled quantity and the mode of the scalation, it should wait
    until you decide when the scalation process is finished, returns an error if required

## Example

We will make a simple filter as an example. This filter will be called `chaos` and will apply some sort
of chaos monkey of scalation process. The filter will stop the chain and not scale (return the current quantity as the new quantity) 50% of the times.

file `chaos.go`:

```golang
package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/themotion/ladder/autoscaler/filter"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
)

const (
	chaosRegName = "chaos"
)

func init() {
	filter.Register(chaosRegName, filter.CreatorFunc(func(ctx context.Context, opts map[string]interface{}) (filter.Filterer, error) {
		return NewChaos(ctx, opts)
	}))
}

// Chaos will drop randomly any scalup or scaledown
type Chaos struct {
	ctx context.Context
	log *log.Log // custom logger
}

// NewChaos returns a new chaos filter
func NewChaos(ctx context.Context, _ map[string]interface{}) (*Chaos, error) {
	asName, ok := ctx.Value("autoscaler").(string)
	if !ok {
		asName = "unknown"
	}
	logger := log.WithFields(log.Fields{
		"autoscaler": asName,
		"kind":       "filterer",
		"name":       chaosRegName,
	})

	return &Chaos{
		ctx: ctx,
		log: logger,
	}, nil
}

// Filter will drop randomly scaling stuff and break the chain
func (c Chaos) Filter(ctx context.Context, currentQ, newQ types.Quantity) (types.Quantity, bool, error) {
	var brk bool
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if r.Intn(100)%2 == 0 {
		newQ = currentQ
		brk = true
		c.log.Warningf("Chaos applied!, breaking the chain and setting to current")
	}
	return newQ, brk, nil
}
```

We start from top to down:

`init` function executes automatically when the plugin is laoded, in this function we register our filter
using ladder filter register helper `filter.Register` with the filter name `chaos`, we register a `filter.CreatorFunc` that is a function
wrapping the creation of our filter object.

`NewChaos` creates the filter instance, it receives the context, from there we get the autoscaling name to
set up correctly on the logger, and it receives the `config` map that we define on the settings, in this
case, this filter doesn't have any options so, we ignore this parameter.

At last `Filter` will apply chaos when a random number is multiple of 2 (50%).

We need to compile the plugin against the version of Ladder that will run so we get that Ladder version and
do `go build -buildmode=plugin -o {DST_PATH}/chaos.so ./chaos.go` and we only need to set up on our `ladder.cfg` file to load that shared lib (.so). Example:

```yaml
global:
  warmup: 10s
  plugins:
    - myplugins/chaos.so

autoscaler_files:
  - "cfg-autoscalers/*.yml"
```

and use it on the autoscalers:

```yaml
...
  filters:
    - kind: chaos
...
```

Check a full ladder plugin project as a further example [here](https://github.com/slok/ladder-plugin-example)

{{< note title="Note" >}}
Plugins file names (f.e `myplugin.go`) can't be the same, See [Go issue](https://github.com/golang/go/issues/19004)
{{< /note >}}

{{< note title="Note" >}}
Ladder plugins should be compiled against the same libc used to compile Ladder. for example
if you use Ladder on themotion docker image (it uses alpine with musl-libc, your plugin should be compiled against musl libc instead of glibc
{{< /note >}}