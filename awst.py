#!/Users/gopukrishnan/awst_env/bin/python3

import argparse



try:
    from prettytable import PrettyTable
except ImportError:
    print("prettytable Module not found. Please run 'pip3 install prettytable'")

try:
    from colorama import Fore, Style
except ImportError:
    print("colorama Module not found. Please run 'pip3 install colorama'")

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
    else:
        print("Unreccognized command: " + unknown_args[0])

if __name__ == "__main__":
    main()
