package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsEc2TrafficMirrorTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TrafficMirrorTargetCreate,
		Read:   resourceAwsEc2TrafficMirrorTargetRead,
		Delete: resourceAwsEc2TrafficMirrorTargetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"network_interface_id",
					"network_load_balancer_arn",
				},
			},
			"network_load_balancer_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"network_interface_id",
					"network_load_balancer_arn",
				},
			},
		},
	}
}

func resourceAwsEc2TrafficMirrorTargetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.CreateTrafficMirrorTargetInput{}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_interface_id"); ok {
		input.NetworkInterfaceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_load_balancer_arn"); ok {
		input.NetworkLoadBalancerArn = aws.String(v.(string))
	}

	out, err := conn.CreateTrafficMirrorTarget(input)
	if err != nil {
		return fmt.Errorf("Error creating traffic mirror target %v", err)
	}

	d.SetId(*out.TrafficMirrorTarget.TrafficMirrorTargetId)

	return resourceAwsEc2TrafficMirrorTargetRead(d, meta)
}

func resourceAwsEc2TrafficMirrorTargetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	targetId := d.Id()
	input := &ec2.DescribeTrafficMirrorTargetsInput{
		TrafficMirrorTargetIds: []*string{&targetId},
	}

	out, err := conn.DescribeTrafficMirrorTargets(input)
	if isAWSErr(err, "InvalidTrafficMirrorTargetId.NotFound", "") {
		log.Printf("[WARN] EC2 Traffic Mirror Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing EC2 Traffic Mirror Target (%s): %w", targetId, err)
	}

	if nil == out || 0 == len(out.TrafficMirrorTargets) {
		log.Printf("[WARN] EC2 Traffic Mirror Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	target := out.TrafficMirrorTargets[0]
	d.Set("description", target.Description)
	d.Set("network_interface_id", target.NetworkInterfaceId)
	d.Set("network_load_balancer_arn", target.NetworkLoadBalancerArn)

	return nil
}

func resourceAwsEc2TrafficMirrorTargetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	targetId := d.Id()
	input := &ec2.DeleteTrafficMirrorTargetInput{
		TrafficMirrorTargetId: &targetId,
	}

	_, err := conn.DeleteTrafficMirrorTarget(input)
	if nil != err {
		return fmt.Errorf("error deleting EC2 Traffic Mirror Target (%s): %w", targetId, err)
	}

	return nil
}
