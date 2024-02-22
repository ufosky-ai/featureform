import featureform.resources
from typing import Union, Optional

from .constants import NO_RECORD_LIMIT
from .register import (
    ResourceClient,
    SourceRegistrar,
    SubscriptableTransformation,
    FeatureColumnResource,
)
from .serving import ServingClient
from .enums import ResourceType


class Client(ResourceClient, ServingClient):
    """
    Client for interacting with Featureform APIs (resources and serving)

    **Using the Client:**
    ```py title="definitions.py"
    import featureform as ff
    from featureform import Client

    client = Client()

    # Example 1: Get a registered provider
    redis = client.get_provider("redis-quickstart")

    # Example 2: Compute a dataframe from a registered source
    transactions_df = client.dataframe("transactions", "quickstart")
    ```
    """

    def __init__(
        self, host=None, local=False, insecure=False, cert_path=None, dry_run=False
    ):
        if local:
            raise Exception(
                "Local mode is not supported in this version. Use featureform <= 1.12.0 for localmode"
            )

        if host is not None:
            self._validate_host(host)

        ResourceClient.__init__(
            self,
            host=host,
            local=local,
            insecure=insecure,
            cert_path=cert_path,
            dry_run=dry_run,
        )
        # Given both ResourceClient and ServingClient are instantiated together, if dry_run is True, then
        # the ServingClient cannot be instantiated due to a conflict the local and host arguments.
        if not dry_run:
            ServingClient.__init__(
                self, host=host, local=local, insecure=insecure, cert_path=cert_path
            )

    def dataframe(
        self,
        source: Union[SourceRegistrar, SubscriptableTransformation, str],
        variant: Optional[str] = None,
        limit=NO_RECORD_LIMIT,
        asynchronous=False,
        verbose=False,
    ):
        """
        Return a dataframe from a registered source or transformation

        **Example:**
        ```py title="definitions.py"
        transactions_df = client.dataframe("transactions", "quickstart")

        avg_user_transaction_df = transactions_df.groupby("CustomerID")["TransactionAmount"].mean()
        ```

        Args:
            source (Union[SourceRegistrar, SubscriptableTransformation, str]): The source or transformation to compute the dataframe from
            variant (str): The source variant; can't be None if source is a string
            limit (int): The maximum number of records to return; defaults to NO_RECORD_LIMIT
            asynchronous (bool): Flag to determine whether the client should wait for resources to be in either a READY or FAILED state before returning. Defaults to False to ensure that newly registered resources are in a READY state prior to serving them as dataframes.

        Returns:
            df (pandas.DataFrame): The dataframe computed from the source or transformation

        """
        self.apply(asynchronous=asynchronous, verbose=verbose)
        if isinstance(source, (SourceRegistrar, SubscriptableTransformation)):
            name, variant = source.name_variant()
        elif isinstance(source, str):
            name = source
            if variant is None:
                raise ValueError("variant must be specified if source is a string")
            if variant == "":
                raise ValueError("variant cannot be an empty string")
        else:
            raise ValueError(
                f"source must be of type SourceRegistrar, SubscriptableTransformation or str, not {type(source)}\n"
                "use client.dataframe(name, variant) or client.dataframe(source) or client.dataframe(transformation)"
            )
        return self.impl._get_source_as_df(name, variant, limit)

    def nearest(self, feature, vector, k):
        """
        Query the K nearest neighbors of a provider vector in the index of a registered feature variant

        **Example:**

        ```py title="definitions.py"
        # Get the 5 nearest neighbors of the vector [0.1, 0.2, 0.3] in the index of the feature "my_feature" with variant "my_variant"
        nearest_neighbors = client.nearest("my_feature", "my_variant", [0.1, 0.2, 0.3], 5)
        print(nearest_neighbors) # prints a list of entities (e.g. ["entity1", "entity2", "entity3", "entity4", "entity5"])
        ```

        Args:
            feature (Union[FeatureColumnResource, tuple(str, str)]): Feature object or tuple of Feature name and variant
            vector (List[float]): Query vector
            k (int): Number of nearest neighbors to return

        """
        if isinstance(feature, tuple):
            name, variant = feature
        elif isinstance(feature, FeatureColumnResource):
            name = feature.name
            variant = feature.variant
        else:
            raise Exception(
                f"the feature '{feature}' of type '{type(feature)}' is not support."
                "Feature must be a tuple of (name, variant) or a FeatureColumnResource"
            )

        if k < 1:
            raise ValueError("k must be a positive integer")
        return self.impl.nearest(name, variant, vector, k)

    def location(
        self,
        source: Union[SourceRegistrar, SubscriptableTransformation, str],
        variant: Optional[str] = None,
        resource_type: Optional[ResourceType] = None,
    ):
        """
        Returns the location of a registered resource. For SQL resources, it will return the table name
        and for file resources, it will return the file path.

        **Example:**
        ```py title="definitions.py"
        transaction_location = client.location("transactions", "quickstart", ff.SOURCE)
        ```

        Args:
            source (Union[SourceRegistrar, SubscriptableTransformation, str]): The source or transformation to compute the dataframe from
            variant (str): The source variant; can't be None if source is a string
            resource_type (ResourceType): The type of resource; can be one of ff.SOURCE, ff.FEATURE, ff.LABEL, or ff.TRAINING_SET
        """
        if isinstance(source, (SourceRegistrar, SubscriptableTransformation)):
            name, variant = source.name_variant()
            resource_type = ResourceType.SOURCE
        elif isinstance(source, featureform.resources.TrainingSetVariant):
            name = source.name
            variant = source.variant
            resource_type = ResourceType.TRAINING_SET
        elif isinstance(source, str):
            name = source
            if variant is None:
                raise ValueError("variant must be specified if source is a string")
            if variant == "":
                raise ValueError("variant cannot be an empty string")

            if resource_type is None:
                raise ValueError(
                    "resource_type must be specified if source is a string"
                )

        else:
            raise ValueError(
                f"source must be of type SourceRegistrar, SubscriptableTransformation or str, not {type(resource)}\n"
                "use client.dataframe(name, variant) or client.dataframe(source) or client.dataframe(transformation)"
            )

        location = self.impl.location(name, variant, resource_type)
        return location

    def close(self):
        """
        Closes the client, closes channel for hosted mode
        """
        self.impl.close()

    def add_trigger(self, trigger_name, resource_name, resource_variant):
        """
        Add a trigger to a resource after creation

        **Example:**
        ```py title="definitions.py"
        client.add_trigger("my_trigger", "my_resource", "my_variant")
        ```

        Args:
            trigger_name (str): The name of the trigger
            resource_name (str): The name of the resource
            resource_variant (str): The variant of the resource
        """
        self.impl.add_trigger(trigger_name, resource_name, resource_variant)
        # raise NotImplementedError("This method is not yet implemented")
    
    def remove_trigger(self, trigger_name, resource_name, resource_variant):
        """
        Remove a trigger from a resource

        **Example:**
        ```py title="definitions.py"
        client.remove_trigger("my_trigger", "my_resource", "my_variant")
        ```

        Args:
            trigger_name (str): The name of the trigger
            resource_name (str): The name of the resource
            resource_variant (str): The variant of the resource
        """
        # self.impl.remove_trigger(trigger_name, resource_name, resource_variant)
        raise NotImplementedError("This method is not yet implemented")
    
    def update_trigger(self, trigger_name, schedule):
        """
        Update a trigger
        """
        # self.impl.update_trigger(trigger_name, schedule)
        raise NotImplementedError("This method is not yet implemented")
    
    def delete_trigger(self, trigger_name):
        """
        Delete a trigger
        """
        # self.impl.delete_trigger(trigger_name)
        raise NotImplementedError("This method is not yet implemented")



    @staticmethod
    def _validate_host(host):
        if host.startswith("http://") or host.startswith("https://"):
            raise ValueError("Invalid Host: Host should not contain http or https.")

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        self.close()
