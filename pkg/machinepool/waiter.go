package machinepool

import (
	"context"
	"fmt"
	"time"

	tm "github.com/buger/goterm"
	"github.com/gdamore/tcell/v2"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/rivo/tview"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	clusterapiExp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
)

const (
	waitTime = 40 * 60 * time.Second
)

func ListAll(kubeConf *rest.Config) ([]clusterapiExp.MachinePool, error) {
	mps, err := NewForConfig(kubeConf)
	if err != nil {
		return nil, err
	}

	client := mps.MachinePools("")
	l, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return l.Items, nil
}

type MachinePoolWaiter interface {
	Init()
	Check(mp *clusterapiExp.MachinePool) bool
}

type machinePoolWaiterClient struct {
	pools *clusterapiExp.MachinePoolList
	phase map[string]clusterapiExp.MachinePoolPhase
	app   *tview.Application
	table *tview.Table
}

func (c *machinePoolWaiterClient) Init() {
	c.phase = make(map[string]clusterapiExp.MachinePoolPhase)
	for _, mp := range c.pools.Items {
		c.phase[mp.Name] = findReadiness(&mp)
	}

	app := tview.NewApplication()
	c.app = app
	table := tview.NewTable().
		SetBorders(true).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
	})
	c.table = table
}

// function that will update the table with the current status of the machine pools
// the table has 2 columns, the first one is the name of the machine pool and the second one is the phase
func (c *machinePoolWaiterClient) UpdateTable() {
	c.table.Clear()
	headers := []string{"Machine Pool", "Phase"}
	for i, header := range headers {
		c.table.SetCell(0, i, tview.NewTableCell(header).SetTextColor(tcell.ColorYellow))
	}
	for i, mp := range c.pools.Items {
		name := mp.Name
		phase := string(c.phase[name])
		c.table.SetCell(i+1, 0, tview.NewTableCell(name))
		c.table.SetCell(i+1, 1, tview.NewTableCell(phase))
	}
}

func (c *machinePoolWaiterClient) Check(mp *clusterapiExp.MachinePool) bool {
	c.phase[mp.Name] = findReadiness(mp)

	c.UpdateTable()
	c.app.Draw()

	return allTrue(c.phase)
	// return false
}

// function that check if all values in map[string]bool are true
func allTrue(m map[string]clusterapiExp.MachinePoolPhase) bool {
	for _, v := range m {
		if v != clusterapiExp.MachinePoolPhaseRunning {
			return false
		}
	}
	return true
}

func AllWaiter(kubeConf *rest.Config, namespace string, clusterName string, timeout func() error) error {
	conf := config.Read()
	ctx := context.Background()
	mps, err := NewForConfig(kubeConf)
	if err != nil {
		return err
	}

	label := &metav1.LabelSelector{MatchLabels: map[string]string{"cluster.x-k8s.io/cluster-name": clusterName}}

	client := mps.MachinePools(conf.Namespace(namespace))
	pools, err := client.List(ctx, metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(label)})
	if err != nil {
		return err
	}
	if len(pools.Items) == 0 {
		return fmt.Errorf("No machine pools found for cluster %s", clusterName)
	}

	waitClient := &machinePoolWaiterClient{pools: pools}

	waitClient.Init()

	go func() {
		if err := waitClient.app.SetRoot(waitClient.table, true).SetFocus(waitClient.table).Run(); err != nil {
			panic(err)
		}
	}()

	if ready := waitClient.Check(&pools.Items[0]); ready {
		return err
	}

	watcher, err := WatchNamespace(ctx, client, metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(label)})
	if err != nil {
		return err
	}

	ch := watcher.ResultChan()
	for {
		select {
		case event := <-ch:
			// tm.Clear()
			mp, ok := event.Object.(*clusterapiExp.MachinePool)
			if !ok {
				return fmt.Errorf("Failed to parse watch event")
			}

			if stop := waitClient.Check(mp); stop {
				waitClient.app.Stop()
				return nil
			}
		case <-time.After(waitTime):
			if err := timeout(); err != nil {
				return err
			}
		}
	}
}

func Waiter(kubeConf *rest.Config, namespace string, name string, mpFunc func(mp *clusterapiExp.MachinePool) (bool, error), timeout func() error) error {
	conf := config.Read()
	ctx := context.Background()
	mps, err := NewForConfig(kubeConf)
	if err != nil {
		return err
	}

	client := mps.MachinePools(conf.Namespace(namespace))
	mp, err := client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	tm.Clear()
	if ready, err := mpFunc(mp); ready || err != nil {
		return err
	}

	watcher, err := WatchNamespace(ctx, client, metav1.ListOptions{})
	if err != nil {
		return err
	}

	ch := watcher.ResultChan()
	for {
		select {
		case event := <-ch:
			tm.Clear()
			mp, ok := event.Object.(*clusterapiExp.MachinePool)
			if !ok {
				return fmt.Errorf("Failed to parse watch event")
			}

			if stop, err := mpFunc(mp); stop || err != nil {
				return err
			}
		case <-time.After(waitTime):
			if err := timeout(); err != nil {
				return err
			}
		}
	}
}

func SilentWait(kubeConf *rest.Config, namespace string, name string) error {
	timeout := func() error {
		return fmt.Errorf("Failed to become ready after 40 minutes, try running `plural cluster mpwatch %s %s` to get an idea where to debug", namespace, name)
	}

	return Waiter(kubeConf, namespace, name, func(mp *clusterapiExp.MachinePool) (bool, error) {
		phase := findReadiness(mp)
		if phase == clusterapiExp.MachinePoolPhaseRunning {
			fmt.Printf("MachinePool %s is finally ready!", name)
			return true, nil
		}
		return false, nil
	}, timeout)
}

func Wait(kubeConf *rest.Config, namespace string, name string) error {
	timeout := func() error {
		return fmt.Errorf("Failed to become ready after 40 minutes, try running `plural cluster mpwatch %s %s` to get an idea where to debug", namespace, name)
	}

	return Waiter(kubeConf, namespace, name, func(mp *clusterapiExp.MachinePool) (bool, error) {
		tm.MoveCursor(1, 1)
		ready := Ready(mp)
		Flush()
		return ready, nil
	}, timeout)
}

func WaitAll(kubeConf *rest.Config, namespace string, clusterName string) error {
	timeout := func() error {
		return fmt.Errorf("Failed to become ready after 40 minutes, try running `plural cluster mpwatch %s %s` to get an idea where to debug", namespace, clusterName)
	}

	return AllWaiter(kubeConf, namespace, clusterName, timeout)
}
