
import argparse

try:
    from prettytable import PrettyTable
except ImportError:
    print("prettytable Module not found. Please run 'pip3 install prettytable'")

try:
    from colorama import Fore, Style
except ImportError:
    print("colorama Module not found. Please run 'pip3 install colorama'")



def describeEc2(session,instance_id=None):
    ec2 = session.resource('ec2')
    instances = list(ec2.instances.filter(InstanceIds=[instance_id]))

    # Check if there are any instances with the specified ID
    if not instances:
        print(f"No instances found with ID: {instance_id}")
        return

    # Extract the instance details
    instance = instances[0]

    # Create a PrettyTable to display the information
    table = PrettyTable()
    table.field_names = ['Parameter', 'Value']

    instance_status = instance.state['Name']

    if instance_status == "stopped":
        instance_status = f"{Fore.RED}{instance_status}{Style.RESET_ALL}"

    elif instance_status == "terminated":
        instance_status = f"{Fore.LIGHTBLUE_EX}{instance_status}{Style.RESET_ALL}"

    elif instance_status == "pending":
        instance_status = f"{Fore.YELLOW}{instance_status}{Style.RESET_ALL}"

    elif instance_status == "running":
        instance_status = f"{Fore.GREEN}{instance_status}{Style.RESET_ALL}"

    lifecycle = instance.instance_lifecycle if hasattr(instance, 'instance_lifecycle') else "N/A"

    # Add instance details to the table
    table.add_row(['Instance ID', instance.id])
    table.add_row(['Instance Name', instance.tags[0]['Value'] if instance.tags else '-'])
    table.add_row(['Instance Type', instances[0].instance_type])
    table.add_row(['Public IP', instance.public_ip_address])
    table.add_row(['Private IP', instance.private_ip_address])
    table.add_row(['Status', instance_status])
    table.add_row(['Lifecycle', lifecycle])
    table.add_row(['Instance Type', instance.instance_type])
    table.add_row(['Availability Zone', instance.placement['AvailabilityZone']])
    table.add_row(['Key Name', instance.key_name])
    table.add_row(['AMI Id', instance.image_id])
    table.add_row(['Security Groups', ', '.join([group['GroupName'] for group in instance.security_groups])])

    # Print the formatted table
    print(table)
        # Display attached volume information
    volumes = list(instance.volumes.all())
    
    if volumes:
        volume_table = PrettyTable()
        volume_table.field_names = ['Volume ID', 'Size (GiB)', 'Device Name', 'Volume Type', 'State']
        
        for volume in volumes:
            volume_table.add_row([
                volume.id,
                volume.size,
                volume.attachments[0]['Device'] if volume.attachments else '',
                volume.volume_type,
                volume.state
            ])
        
        print("\nAttached Volumes:")
        print(volume_table)
    else:
        print("\nNo attached volumes for this instance.")






def listEc2(session):
    # Create an EC2 resource client
    ec2 = session.resource('ec2')
    table = PrettyTable(["Name", "ID", "Status", "Public IP", "Private IP"])
    # Iterate through all EC2 instances and retrieve the desired information
    for instance in ec2.instances.all():
        instance_name = '-'
        
        try:
            for tag in instance.tags:
                if tag['Key'] == 'Name':
                    instance_name = tag['Value']
        except:
            pass

        instance_id = instance.id
        instance_status = instance.state['Name']
        public_ip = instance.public_ip_address if instance.public_ip_address else f"{Fore.YELLOW}N/A{Style.RESET_ALL}"
        private_ip = instance.private_ip_address if instance.private_ip_address else f"{Fore.YELLOW}N/A{Style.RESET_ALL}"


        if instance_status == "stopped":
            instance_status = f"{Fore.RED}{instance_status}{Style.RESET_ALL}"

        elif instance_status == "terminated":
            instance_status = f"{Fore.LIGHTBLUE_EX}{instance_status}{Style.RESET_ALL}"

        elif instance_status == "pending":
            instance_status = f"{Fore.YELLOW}{instance_status}{Style.RESET_ALL}"

        # Add the instance information to the table
        table.add_row([instance_name, instance_id,instance_status, public_ip, private_ip])

    print(table)


def main():
    parser = argparse.ArgumentParser(description='Fetch AWS EC2 instances list')
    parser.add_argument('--region', type=str, help='AWS region name', required=True)
    parser.add_argument('--profile', type=str, help='AWS profile name', required=False)
    parser.add_argument('--id', type=str, help='EC2 instance id', required=False)

    args, unknown_args = parser.parse_known_args()


    aws_profile_name = args.profile
    aws_region = args.region
    try:
        import boto3  
    except ImportError:
        print("Boto3 Module not found. Please run 'pip3 install boto3'")


    # Create a session using the specified AWS profile
    session = boto3.Session(profile_name=aws_profile_name, region_name=aws_region)

    if ("--list-ec2" in unknown_args):
        listEc2(session)
    elif ("--describe-ec2" in unknown_args):

        if (args.id):
            describeEc2(session,args.id)
        else:
            print("Instance id required\nUSAGE: --id <instance-id>")
    else:
        print("Unreccognized command: " + unknown_args[0])

if __name__ == "__main__":
    main()
