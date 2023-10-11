package main

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/logging"

	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"

	"github.com/23technologies/xp-fn-rndstr/input/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Function returns whatever response you ask it to.
type Function struct {
	fnv1beta1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1beta1.RunFunctionRequest) (*fnv1beta1.RunFunctionResponse, error) {
	f.log.Info("Running Function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	in := &v1beta1.RandString{}
	if err := request.GetInput(req, in); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from %T", req))
		return rsp, nil
	}

	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		return nil, err
	}

	observed, err := request.GetObservedComposedResources(req)
	if err != nil {
		return nil, err
	}

	b := make([]byte, in.Cfg.RandStr.Length)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	randString := fmt.Sprintf("%X", b)

	for _, obj := range in.Cfg.Objs {
		if observed[resource.Name(obj.Name)].Resource != nil {
			observedPaved, err := fieldpath.PaveObject(observed[resource.Name(obj.Name)].Resource)
			if err != nil {
				return nil, err
			}
			randString, err = observedPaved.GetString(obj.FieldPath)
			if err != nil {
				return nil, err
			}
		}
		patchFieldValueToObject(obj.FieldPath, randString, desired[resource.Name(obj.Name)].Resource)
	}

	response.SetDesiredComposedResources(rsp, desired)
	return rsp, nil
}

func patchFieldValueToObject(fieldPath string, value any, to runtime.Object) error {
	paved, err := fieldpath.PaveObject(to)
	if err != nil {
		return err
	}

	if err := paved.SetValue(fieldPath, value); err != nil {
		return err
	}

	return runtime.DefaultUnstructuredConverter.FromUnstructured(paved.UnstructuredContent(), to)
}
