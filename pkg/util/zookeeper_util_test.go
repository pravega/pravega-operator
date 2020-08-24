/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */
package util

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("zookeeperutil", func() {
	Context("DeleteAllZnodes", func() {
		var err error
		BeforeEach(func() {
			err = DeleteAllZnodes("zookeeper-client:2181", "pravega")
		})
		It("should not be nil", func() {
			Ω(err).ShouldNot(BeNil())
		})
	})
})
