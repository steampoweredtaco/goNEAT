package network

import (
	"github.com/steampoweredtaco/goNEAT/v3/neat/math"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func buildPlainNetwork() *Network {
	allNodes := []*NNode{
		NewNNode(1, InputNeuron),
		NewNNode(2, InputNeuron),
		NewNNode(3, BiasNeuron),
		NewNNode(7, OutputNeuron),
		NewNNode(8, OutputNeuron),
	}

	// OUTPUT 7
	allNodes[3].ConnectFrom(allNodes[1], 7.0)
	allNodes[3].ConnectFrom(allNodes[2], 4.5)
	// OUTPUT 8
	allNodes[4].ConnectFrom(allNodes[3], 13.0)

	return NewNetwork(allNodes[0:3], allNodes[3:5], allNodes, 0)
}

func buildDisconnectedNetwork() *Network {
	allNodes := []*NNode{
		NewNNode(1, InputNeuron),
		NewNNode(2, InputNeuron),
		NewNNode(3, BiasNeuron),
		NewNNode(4, HiddenNeuron),
		NewNNode(5, HiddenNeuron),
		NewNNode(6, HiddenNeuron),
		NewNNode(7, OutputNeuron),
		NewNNode(8, OutputNeuron),
	}
	return NewNetwork(allNodes[0:3], allNodes[6:8], allNodes, 0)
}

func buildNetwork() *Network {
	allNodes := []*NNode{
		NewNNode(1, InputNeuron),
		NewNNode(2, InputNeuron),
		NewNNode(3, BiasNeuron),
		NewNNode(4, HiddenNeuron),
		NewNNode(5, HiddenNeuron),
		NewNNode(6, HiddenNeuron),
		NewNNode(7, OutputNeuron),
		NewNNode(8, OutputNeuron),
	}

	// HIDDEN 4
	allNodes[3].ConnectFrom(allNodes[0], 15.0)
	allNodes[3].ConnectFrom(allNodes[1], 10.0)
	// HIDDEN 5
	allNodes[4].ConnectFrom(allNodes[1], 5.0)
	allNodes[4].ConnectFrom(allNodes[2], 1.0)
	// HIDDEN 6
	allNodes[5].ConnectFrom(allNodes[4], 17.0)
	// OUTPUT 7
	allNodes[6].ConnectFrom(allNodes[3], 7.0)
	allNodes[6].ConnectFrom(allNodes[5], 4.5)
	// OUTPUT 8
	allNodes[7].ConnectFrom(allNodes[5], 13.0)

	return NewNetwork(allNodes[0:3], allNodes[6:8], allNodes, 0)
}

func buildModularNetwork() *Network {
	allNodes := []*NNode{
		NewNNode(1, InputNeuron),
		NewNNode(2, InputNeuron),
		NewNNode(3, BiasNeuron),
		NewNNode(4, HiddenNeuron),
		NewNNode(5, HiddenNeuron),
		NewNNode(7, HiddenNeuron),
		NewNNode(8, OutputNeuron),
		NewNNode(9, OutputNeuron),
	}
	controlNodes := []*NNode{
		NewNNode(6, HiddenNeuron),
	}
	// HIDDEN 6 - control node
	controlNodes[0].ActivationType = math.MultiplyModuleActivation
	controlNodes[0].AddIncoming(allNodes[3], 1.0)
	controlNodes[0].AddIncoming(allNodes[4], 1.0)
	controlNodes[0].AddOutgoing(allNodes[5], 1.0)

	// HIDDEN 4
	allNodes[3].ActivationType = math.LinearActivation
	allNodes[3].ConnectFrom(allNodes[0], 15.0)
	allNodes[3].ConnectFrom(allNodes[2], 10.0)
	// HIDDEN 5
	allNodes[4].ActivationType = math.LinearActivation
	allNodes[4].ConnectFrom(allNodes[1], 5.0)
	allNodes[4].ConnectFrom(allNodes[2], 1.0)

	// HIDDEN 7
	allNodes[5].ActivationType = math.NullActivation

	// OUTPUT 8
	allNodes[6].ConnectFrom(allNodes[5], 4.5)
	allNodes[6].ActivationType = math.LinearActivation
	// OUTPUT 9
	allNodes[7].ConnectFrom(allNodes[5], 13.0)
	allNodes[7].ActivationType = math.LinearActivation

	return NewModularNetwork(allNodes[0:3], allNodes[6:8], allNodes, controlNodes, 0)
}

func TestModularNetwork_Activate(t *testing.T) {
	net := buildModularNetwork()

	data := []float64{1.0, 2.0, 1.0}
	err := net.LoadSensors(data)
	require.NoError(t, err, "failed to load sensors")

	for i := 0; i < 5; i++ {
		res, err := net.Activate()
		require.NoError(t, err, "error when do activation at: %d", i)
		require.True(t, res, "failed to activate at: %d", i)
	}
	assert.Equal(t, 1237.5, net.Outputs[0].Activation)
	assert.Equal(t, 3575.0, net.Outputs[1].Activation)
}

// Tests MaxActivationDepth for simple network
func TestNetwork_MaxActivationDepth_Simple(t *testing.T) {
	net := buildNetwork()

	depth, err := net.MaxActivationDepth()
	assert.NoError(t, err, "failed to calculate max depth")
	assert.Equal(t, 4, depth)

	logNetworkActivationPath(net, t)
}

func TestNetwork_MaxActivationDepth_Modular(t *testing.T) {
	net := buildModularNetwork()

	depth, err := net.MaxActivationDepth()
	assert.NoError(t, err, "failed to calculate max depth")
	assert.Equal(t, 5, depth)

	logNetworkActivationPath(net, t)
}

func TestNetwork_MaxActivationDepth_No_Hidden_or_Control(t *testing.T) {
	net := buildPlainNetwork()
	depth, err := net.MaxActivationDepth()
	assert.NoError(t, err, "failed to calculate max depth")
	assert.Equal(t, 1, depth)
}

func TestNetwork_MaxActivationDepthFast_Simple(t *testing.T) {
	net := buildNetwork()

	depth, err := net.MaxActivationDepthFast(0)
	assert.NoError(t, err, "failed to calculate max depth")
	assert.Equal(t, 4, depth)

	logNetworkActivationPath(net, t)
}

func TestNetwork_MaxActivationDepthFast_Simple_WithMaxLimitError(t *testing.T) {
	net := buildNetwork()

	maxDepth := 2
	depth, err := net.MaxActivationDepthFast(2)
	assert.EqualError(t, err, ErrMaximalNetDepthExceeded.Error())
	assert.Equal(t, maxDepth, depth)
}

func TestNetwork_MaxActivationDepthFast_Modular(t *testing.T) {
	net := buildModularNetwork()

	_, err := net.MaxActivationDepthFast(0)
	assert.Error(t, err, "error expected")
}

// Tests Network OutputIsOff
func TestNetwork_OutputIsOff(t *testing.T) {
	net := buildNetwork()

	res := net.OutputIsOff()
	assert.True(t, res)
}

// Tests Network Activate
func TestNetwork_Activate(t *testing.T) {
	net := buildNetwork()

	res, err := net.Activate()
	require.NoError(t, err, "error when do activation at")
	require.True(t, res, "failed to activate at")

	// check activation
	for i, node := range net.AllNodes() {
		if node.IsNeuron() {
			require.NotZero(t, node.ActivationsCount, "ActivationsCount not set at: %d", i)
			require.NotZero(t, node.Activation, "Activation not set at: %d", i)

			// Check activation and time delayed activation
			require.NotZero(t, node.GetActiveOut(), "GetActiveOut not set at: %d", i)
		}
	}
}

func TestNetwork_ForwardSteps(t *testing.T) {
	net := buildNetwork()

	// test normal activation
	res, err := net.ForwardSteps(3)
	assert.NoError(t, err)
	assert.True(t, res)

	expectedOuts := []float64{1.0, 1.0}
	assert.EqualValues(t, expectedOuts, net.ReadOutputs())

	// test zero steps
	res, err = net.ForwardSteps(0)
	assert.EqualError(t, err, ErrZeroActivationStepsRequested.Error())
	assert.False(t, res)
}

func TestNetwork_RecursiveSteps(t *testing.T) {
	net := buildNetwork()

	data := []float64{0.5, 0.0, 1.5}
	err := net.LoadSensors(data)
	require.NoError(t, err, "failed to load sensors")

	relaxed, err := net.RecursiveSteps()
	assert.NoError(t, err)
	assert.True(t, relaxed)

	logNetworkActivationPath(net, t)

	expectedOuts := []float64{1.0, 1.0}
	assert.EqualValues(t, expectedOuts, net.ReadOutputs())
}

func TestNetwork_ForwardSteps_disconnected(t *testing.T) {
	net := buildDisconnectedNetwork()

	res, err := net.ForwardSteps(3)
	assert.EqualError(t, err, ErrNetExceededMaxActivationAttempts.Error())
	assert.False(t, res)
}

// Test Network LoadSensors
func TestNetwork_LoadSensors(t *testing.T) {
	net := buildNetwork()

	sensors := []float64{1.0, 3.4, 5.6}

	err := net.LoadSensors(sensors)
	require.NoError(t, err, "failed to load sensors")

	counter := 0
	for i, node := range net.AllNodes() {
		if node.IsSensor() {
			assert.Equal(t, sensors[counter], node.Activation, "Sensor value wrong at: %d", i)
			assert.EqualValues(t, 1, node.ActivationsCount, "Sensor activations count wrong at: %d", i)
			counter++
		}
	}
}

// Test Network Flush
func TestNetwork_Flush(t *testing.T) {
	net := buildNetwork()

	// activate and check state
	res, err := net.Activate()
	require.NoError(t, err, "error when do activation at")
	require.True(t, res, "failed to activate at")

	// flush and check
	res, err = net.Flush()
	require.NoError(t, err, "error while trying to flush")
	require.True(t, res, "Network flush failed")

	for i, node := range net.AllNodes() {
		assert.Zero(t, node.ActivationsCount, "at %d", i)
		assert.Zero(t, node.Activation, "at %d", i)

		// Check activation and time delayed activation
		assert.Zero(t, node.GetActiveOut(), "at %d", i)
		assert.Zero(t, node.GetActiveOutTd(), "at %d", i)
	}
}

// Tests Network NodeCount
func TestNetwork_NodeCount(t *testing.T) {
	net := buildNetwork()

	count := net.NodeCount()
	assert.Equal(t, 8, count, "Wrong network's node count")
}

// Tests Network LinkCount
func TestNetwork_LinkCount(t *testing.T) {
	net := buildNetwork()

	count := net.LinkCount()
	assert.Equal(t, 8, count, "Wrong network's link count")
}

// Tests Network IsRecurrent
func TestNetwork_IsRecurrent(t *testing.T) {
	net := buildNetwork()

	nodes := net.AllNodes()
	visited := 0 // the count of times the node was visited
	recur := net.IsRecurrent(nodes[0], nodes[7], &visited, 32)
	assert.False(t, recur, "Network is not recurrent")
	assert.Equal(t, 1, visited)

	// Introduce recurrence
	visited = 0
	nodes[4].AddIncoming(nodes[7], 3.0)
	recur = net.IsRecurrent(nodes[5], nodes[7], &visited, 32)
	assert.True(t, recur, "Network is actually recurrent now")
	assert.Equal(t, 5, visited)
}

// test fast network solver generation
func TestNetwork_FastNetworkSolver(t *testing.T) {
	net := buildModularNetwork()

	solver, err := net.FastNetworkSolver()
	require.NoError(t, err, "failed to create fast network solver")
	require.NotNil(t, solver)

	// check solver structure
	assert.Equal(t, net.NodeCount(), solver.NodeCount(), "wrong number of nodes")
	assert.Equal(t, net.LinkCount(), solver.LinkCount(), "wrong number of links")
}

func TestNetwork_ActivateSteps_zero_activation_steps(t *testing.T) {
	net := buildNetwork()

	res, err := net.ActivateSteps(0)
	assert.EqualError(t, err, ErrZeroActivationStepsRequested.Error())
	assert.False(t, res)
}

func TestNetwork_ActivateSteps_ErrNetExceededMaxActivationAttempts(t *testing.T) {
	net := buildDisconnectedNetwork()

	res, err := net.ActivateSteps(10)
	assert.EqualError(t, err, ErrNetExceededMaxActivationAttempts.Error())
	assert.False(t, res)
}

func TestNetwork_ActivateSteps_maxActivationDepth_disconnected(t *testing.T) {
	net := buildDisconnectedNetwork()

	depth, err := net.maxActivationDepth(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, depth)
}

func TestNetwork_ActivateSteps_maxActivationDepth_negative_cycle(t *testing.T) {
	net := buildNetwork()

	// create negative cycle
	net.allNodes[1].ConnectFrom(net.allNodes[7], -130.0)

	depth, err := net.maxActivationDepth(nil)
	assert.NoError(t, err)
	assert.Equal(t, 3, depth)

	logNetworkActivationPath(net, t)
}

func TestNetwork_ActivateSteps_maxActivationDepth_writeError(t *testing.T) {
	net := buildNetwork()

	errWriter := ErrorWriter(1)
	depth, err := net.maxActivationDepth(&errWriter)
	assert.EqualError(t, err, alwaysErrorText)
	assert.Equal(t, 0, depth)
}

func TestModularNetwork_ControlNodes(t *testing.T) {
	net := buildModularNetwork()

	cNodes := net.ControlNodes()
	assert.NotNil(t, cNodes)
	assert.Len(t, cNodes, len(net.controlNodes))
}

func TestModularNetwork_BaseNodes(t *testing.T) {
	net := buildModularNetwork()

	baseNodes := net.BaseNodes()
	assert.NotNil(t, baseNodes)
	assert.Len(t, baseNodes, len(net.allNodes))
}
