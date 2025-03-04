// Copyright 2021 Upbound Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package billing

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"

	"github.com/upbound/universal-crossplane/internal/meta"
)

const (
	reconcileTimeout = 1 * time.Minute
	syncPeriod       = 1 * time.Minute

	errGetSecret       = "cannot get entitlement secret"
	errGetKubesystemNS = "cannot get kube-system namespace"
	errRegister        = "cannot register entitlement"
	errVerify          = "cannot verify signature"
)

// ReconcilerOption is used to configure the Reconciler.
type ReconcilerOption func(*Reconciler)

// WithRegisterer specifies the Registerer to use.
func WithRegisterer(reg Registerer) ReconcilerOption {
	return func(r *Reconciler) {
		r.entitlement = reg
	}
}

// WithLogger specifies how the Reconciler should log messages.
func WithLogger(log logging.Logger) ReconcilerOption {
	return func(r *Reconciler) {
		r.log = log
	}
}

// WithRecorder specifies how the Reconciler should record Kubernetes events.
func WithRecorder(er event.Recorder) ReconcilerOption {
	return func(r *Reconciler) {
		r.record = er
	}
}

// Reconciler reconciles on entitlement secret.
type Reconciler struct {
	client client.Client
	log    logging.Logger
	record event.Recorder

	entitlement Registerer
}

// NewReconciler returns a new reconciler.
func NewReconciler(mgr manager.Manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		client:      mgr.GetClient(),
		log:         logging.NewNopLogger(),
		record:      event.NewNopRecorder(),
		entitlement: NewNopRegisterer(),
	}

	for _, f := range opts {
		f(r)
	}

	return r
}

// Reconcile reconciles on entitlement secret and registers & verifies that usage
// is valid in given entitlement context.
func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := r.log.WithValues("request", req)

	log.Debug("Reconciling...")
	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	s := &corev1.Secret{}
	nn := types.NamespacedName{Name: meta.SecretNameEntitlement, Namespace: req.Namespace}
	if err := r.client.Get(ctx, nn, s); err != nil {
		return reconcile.Result{}, errors.Wrap(err, errGetSecret)
	}

	kubeNS := &corev1.Namespace{}
	nn = types.NamespacedName{Name: "kube-system"}
	if err := r.client.Get(ctx, nn, kubeNS); err != nil {
		return reconcile.Result{}, errors.Wrap(err, errGetKubesystemNS)
	}
	uid := string(kubeNS.GetUID())

	token, err := r.entitlement.Register(ctx, s, uid)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, errRegister)
	}

	verified, err := r.entitlement.Verify(token, uid)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, errVerify)
	}
	if !verified {
		// TODO(muvaf): There is no action we can take at this point.
		log.Info("entitlement signature is not valid")
		return reconcile.Result{RequeueAfter: syncPeriod}, nil
	}

	log.Info("entitlement has been confirmed")
	return reconcile.Result{}, nil
}
