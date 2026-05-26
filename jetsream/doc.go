// Package jetsream implements a client connected to a Jetsream.
//
// [New] creates a [Feed] that can be connected to any Jetsream instance.
// It takes the Jetstream URL (including the `/subscribe` of the official implementations) and [Options] describing
// what do you want.
//
// [Connect], [Reconnect], [Disconnect] and [Connected] manage the [Feed].
//
// [Listen] returns the channel of [Event] sent by Jetstream.
// You *must* listen and consume everything, otherwise it will be stuck.
//
// [SubscriberSourcedMessage] is a message that can be sent to Jetstream.
// Currently, only [SubscriberOptionsUpdateMsg] exists.
package jetsream
