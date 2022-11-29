/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	customprojectv1 "secret-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateSecretReconciler reconciles a CreateSecret object
type CreateSecretReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	lowerCharSet   = "abcdedfghijklmnopqrst"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialCharSet = "!@#$%&*"
	numberSet      = "0123456789"
	allCharSet     = lowerCharSet + upperCharSet + specialCharSet + numberSet
)

func generateRandomSecret(passwordLength, minSpecialChar, minNum, minUpperCase int) string {
	var password strings.Builder

	//Set special character
	for i := 0; i < minSpecialChar; i++ {
		random := rand.Intn(len(specialCharSet))
		password.WriteString(string(specialCharSet[random]))
	}

	//Set numeric
	for i := 0; i < minNum; i++ {
		random := rand.Intn(len(numberSet))
		password.WriteString(string(numberSet[random]))
	}

	//Set uppercase
	for i := 0; i < minUpperCase; i++ {
		random := rand.Intn(len(upperCharSet))
		password.WriteString(string(upperCharSet[random]))
	}

	remainingLength := passwordLength - minSpecialChar - minNum - minUpperCase
	for i := 0; i < remainingLength; i++ {
		random := rand.Intn(len(allCharSet))
		password.WriteString(string(allCharSet[random]))
	}
	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})
	return string(inRune)
}

const secretFinalizer = "customproject.io/secretfinalizer"

//+kubebuilder:rbac:groups=customproject.io,resources=createsecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=customproject.io,resources=createsecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=customproject.io,resources=createsecrets/finalizers,verbs=update

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups="",resources=secrets/finalizers,verbs=update

func (r *CreateSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var createsecret customprojectv1.CreateSecret
	errc := r.Get(ctx, req.NamespacedName, &createsecret)

	if errc != nil {
		if apierrors.IsNotFound(errc) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Log.Info("CreateSecret Resouce not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errc
	}

	if createsecret.Status.SecretCreationTime == "" {
		log.Log.Info("Creating Secret: " + createsecret.Spec.SecretName + " Namespace: " + createsecret.Namespace)
		passwordlength := createsecret.Spec.SecretLength
		password := generateRandomSecret(passwordlength, int((0.25)*float32(passwordlength)), int((0.25)*float32(passwordlength)), int((0.25)*float32(passwordlength)))
		secret := corev1.Secret{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      createsecret.Spec.SecretName,
				Namespace: createsecret.Namespace,
				Labels: map[string]string{
					"SecretManagedBy": createsecret.Name,
				},
			},
			StringData: map[string]string{"secret": password},
		}
		if err := r.Client.Create(ctx, &secret); err != nil {
			if apierrors.IsAlreadyExists(err) {
				log.Log.Info("Secret named : " + createsecret.Spec.SecretName + " already exists in Namespace: " + createsecret.Namespace + " Ignoring Creation")
			}
			log.Log.Error(err, "Unable to Create secret Object : "+createsecret.Spec.SecretName)
			return ctrl.Result{}, err
		} else {
			log.Log.Info("Secret Created: " + createsecret.Spec.SecretName + " Namespace: " + createsecret.Namespace)
		}
		createsecret.Status = customprojectv1.CreateSecretStatus{
			SecretCreationTime: time.Now().String(),
			SecretLength:       passwordlength,
			SecretName:         createsecret.Spec.SecretName,
		}
		if err := r.Status().Update(ctx, &createsecret); err != nil {
			log.Log.Error(err, "Unable to update Status of CreateSecret object : "+createsecret.Spec.SecretName)
			return ctrl.Result{}, err
		}

	} else if createsecret.Status.SecretLength != createsecret.Spec.SecretLength {
		log.Log.Info("Secrets Length changed. Updating Secret: " + createsecret.Spec.SecretName + " Mamespace: " + createsecret.Namespace)
		passwordlength := createsecret.Spec.SecretLength
		password := generateRandomSecret(passwordlength, int((0.25)*float32(passwordlength)), int((0.25)*float32(passwordlength)), int((0.25)*float32(passwordlength)))
		secret := corev1.Secret{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      createsecret.Spec.SecretName,
				Namespace: createsecret.Namespace,
				Labels: map[string]string{
					"SecretManagedBy": createsecret.Name,
				},
			},
			StringData: map[string]string{"secret": password},
		}
		if err := r.Client.Update(ctx, &secret); err != nil {
			log.Log.Error(err, "Unable to Update secret")
			return ctrl.Result{}, err
		}
		createsecret.Status = customprojectv1.CreateSecretStatus{
			SecretCreationTime: time.Now().String(),
			SecretLength:       passwordlength,
			SecretName:         createsecret.Spec.SecretName,
		}
		if err := r.Status().Update(ctx, &createsecret); err != nil {
			log.Log.Error(err, "Unable to update Status of CreateSecret object : "+createsecret.Spec.SecretName+" ")
			return ctrl.Result{}, err
		}
		log.Log.Info("Secret Updated: " + createsecret.Spec.SecretName + " Namespace: " + createsecret.Namespace)

	}

	// Using finalizers to delete secret object
	if createsecret.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(&createsecret, secretFinalizer) {
			log.Log.Info("Deletion Request Received. Deleting Secret: " + createsecret.Spec.SecretName + " Namespace: " + createsecret.Namespace)
			secret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      createsecret.Spec.SecretName,
					Namespace: createsecret.Namespace,
				},
			}
			err := r.Client.Delete(ctx, &secret)
			if err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(&createsecret, secretFinalizer)
			if err := r.Update(ctx, &createsecret); err != nil {
				return ctrl.Result{}, err
			}
			log.Log.Info("Deletion Request Completed. Secret: " + createsecret.Spec.SecretName + " Namespace: " + createsecret.Namespace)
			return ctrl.Result{}, nil
		}
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(&createsecret, secretFinalizer) {
		controllerutil.AddFinalizer(&createsecret, secretFinalizer)
		err := r.Update(ctx, &createsecret)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CreateSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&customprojectv1.CreateSecret{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
