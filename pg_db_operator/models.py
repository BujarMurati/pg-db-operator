from pydantic import BaseModel, Field


class TargetSecretSpec(BaseModel):
    name: str
    user_name_postfix: str = Field(default="", alias="userNamePostfix")


class PostgresDatabaseSpec(BaseModel):
    name: str
    target_secret: TargetSecretSpec = Field(alias="targetSecret")

    class Config:
        allow_population_by_field_name = True

    @property
    def user(self) -> str:
        return self.name + self.target_secret.user_name_postfix
